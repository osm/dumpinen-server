package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"filippo.io/age"
	"github.com/osm/flen"
	"golang.org/x/crypto/acme/autocert"
)

// isValidPublicFileID is a regular expression that checks that the given data
// is a valid public file id.
var isValidPublicFileID = regexp.MustCompile(`^[a-zA-Z0-9\-]+$`).MatchString

// newPublicFileID generates a new public file id.
func newPublicFileID() string {
	return rndStr(11)
}

// app holds the main structure of this application.
type app struct {
	db          *db
	dataDir     string
	port        string
	maxFileSize int64
	recipient   *age.X25519Recipient
	identity    *age.X25519Identity
}

// newApp returns a new app.
func newApp(db *db, dataDir, port, pubKey, privKey string, maxFileSize int64) (*app, error) {
	app := &app{
		db:          db,
		dataDir:     dataDir,
		port:        port,
		maxFileSize: maxFileSize,
	}

	// Make sure the provided age public and private keys are possible to
	// parse before we procced.
	recipient, err := age.ParseX25519Recipient(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key %q: %v", pubKey, err)
	}
	app.recipient = recipient

	identity, err := age.ParseX25519Identity(privKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %v", err)
	}
	app.identity = identity

	return app, nil
}

func getCert(certManager *autocert.Manager) func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		return certManager.GetCertificate(hello)
	}
}

func main() {
	// Add command flags and parse them.
	cs := flag.String("cs", "", "database connection string")
	dataDir := flag.String("data-dir", "", "data directory for uploaded files")
	letsEncryptDomain := flag.String("lets-encrypt-domain", "", "let's encrypt domain name")
	letsEncryptCertDir := flag.String("lets-encrypt-cert-dir", "certs", "let's encrypt cert dir")
	maxFileSize := flag.Int64("max-file-size", 100000000, "max file size, defaults to 100 MB.")
	port := flag.String("port", "80", "port to listen on, the port is only used if domain is localhost")
	privKey := flag.String("priv-key", "", "private age enryption key")
	pubKey := flag.String("pub-key", "", "public age enryption key")
	flen.SetEnvPrefix("DUMPINEN")
	flen.Parse()

	// Make sure that the server is started with a connection string and
	// that we are able to connect with the provided details.
	if *cs == "" {
		fmt.Fprintf(os.Stderr, "-cs is required\n")
		return
	}
	db, err := newDB(*cs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to initialize a database connection, %w\n", err)
		return
	}

	// Verify that the user has submitted a data dir and that the dir
	// exists.
	if *dataDir == "" {
		fmt.Fprintf(os.Stderr, "-data-dir is required\n")
		return
	}
	if _, err := os.Stat(*dataDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: %s is not a valid directory\n", *dataDir)
		return
	}

	// Make sure that the public and private keys are set
	if *pubKey == "" {
		fmt.Fprintf(os.Stderr, "-pub-key is required and needs to be a valid age public key\n")
		return
	}
	if *privKey == "" {
		fmt.Fprintf(os.Stderr, "-priv-key is required and needs to be a valid age private key\n")
		return
	}

	// Create a new app structure and launch the app.
	app, err := newApp(db, *dataDir, *port, *pubKey, *privKey, *maxFileSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "new app error: %v\n", err)
		return
	}

	// Start the cleaner in a goroutine.
	go app.cleaner()

	// Create the mux and push the router to it.
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.router)

	// If we don't have a let's encrypt domain set we'll just launch the
	// server as a regular HTTP server.
	// But if the lets-encrypt-domin is set we'll instead try to acquire a
	// certificate from let's encrypt. We will also not care about the
	// -port flag in this case, instead we'll listen to port 80 and 443.
	if *letsEncryptDomain == "" {
		if err := http.ListenAndServe(":"+*port, mux); err != nil {
			log.Fatalf("fatal error: %v\n", err)
		}
	} else {
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(*letsEncryptDomain),
			Cache:      autocert.DirCache(*letsEncryptCertDir),
		}
		tlsConfig := certManager.TLSConfig()
		tlsConfig.GetCertificate = func(h *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return certManager.GetCertificate(h)
		}

		server := http.Server{
			Addr:      ":443",
			Handler:   mux,
			TLSConfig: tlsConfig,
		}

		go http.ListenAndServe(":80", certManager.HTTPHandler(nil))
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatalf("fatal error: %v\n", err)
		}
	}
}
