package main

import (
	"crypto/subtle"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

// notFound sets the status code to not found and writes a "not found" message
// to the response writer.
func notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found\r\n"))
}

// internalServerError sets the status code to internal server error and
// writes a "internal server error" message to the response writer.
func internalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("internal server error occured, try again later\r\n"))
}

// unauthorized sets the status code to unauthorized and writes an
// "unauthorized" message to the response writer.
func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted area"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("unauthorized\r\n"))
}

// router handles all incoming requests and forwards them to the correct
// location.
func (a *app) router(w http.ResponseWriter, r *http.Request) {
	// Set content type to text/plain for all responses.
	w.Header().Set("Content-Type", "text/plain")

	// Route the request to the correct handler.
	if r.Method == http.MethodPost {
		a.routePost(w, r)
	} else {
		a.routeGet(w, r)
	}
}

// routePost handles the v1 dump POST request.
func (a *app) routePost(w http.ResponseWriter, r *http.Request) {
	log.Printf("dump post request from %s\n", r.RemoteAddr)

	// Add a file size limit and read the contents and do some error
	// checking.
	r.Body = http.MaxBytesReader(w, r.Body, a.maxFileSize)
	data, err := ioutil.ReadAll(r.Body)

	// When we get an empty body we'll return an error and return.
	if len(data) == 0 {
		log.Printf("dump rejected, empty payload\n")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("empty request payload\r\n"))
		return
	}

	if err != nil {
		if int64(len(data)) >= a.maxFileSize {
			log.Printf("dump rejected, payload too big\n")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("request payload too big\r\n"))
			return
		}
		log.Printf("error on dump: %v\n", err)
		internalServerError(w)
		return
	}

	// Check if the user submitted a deleteAfter query parameter and that
	// it was of a valid format.
	var deleteAfter time.Time
	if da, ok := r.URL.Query()["deleteAfter"]; ok {
		d, err := time.ParseDuration(da[0])
		if err != nil {
			log.Printf("error when parsing delete after duration: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "error: invalid deleteAfter duration\r\n")
			return
		}
		deleteAfter = time.Now().Local().Add(d)
	}

	// Determine content type and generate ids for the file.
	contentType := http.DetectContentType(data)
	filesystemID := newUUID()
	publicID := newPublicFileID()

	// If the request contains basic auth credentials we'll use them to
	// protect the uploaded file.
	var username, password []byte
	if u, p, ok := r.BasicAuth(); ok {
		if username, err = a.encrypt(u); err != nil {
			log.Printf("error when encrypting username, %v\n", err)
			internalServerError(w)
			return
		}
		if password, err = a.encrypt(p); err != nil {
			log.Printf("error when encrypting username, %v\n", err)
			internalServerError(w)
			return
		}
	}

	// Insert the dump into the database.
	err = a.db.insertDump(&dump{
		contentType:  contentType,
		deleteAfter:  deleteAfter,
		filesystemID: filesystemID,
		ipAddress:    r.RemoteAddr,
		password:     &password,
		publicID:     publicID,
		username:     &username,
	})
	if err != nil {
		log.Printf("insert dump error: %v\n", err)
		internalServerError(w)
		return
	}

	// Store the actual contents of the file in the database.
	if err = ioutil.WriteFile(filepath.Join(a.dataDir, filesystemID), data, 0440); err != nil {
		log.Printf("write file error: %v\n", err)
		internalServerError(w)
		a.db.deleteDumpByFilesystemID(filesystemID)
		return
	}

	// Set http status code to 201 and return the URL to the stored file.
	log.Printf("dump stored with public id at %s\n", publicID)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s\r\n", publicID)
}

// routeGet handles the v1 dump GET request.
func (a *app) routeGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("dump get request from %s\n", r.RemoteAddr)

	// Discard the / in the beginning of the path.
	publicID := r.URL.Path[1:]
	if len(publicID) != 11 || !isValidPublicFileID(publicID) {
		notFound(w)
		return
	}

	// Query the database for information about the requested file.
	dump, err := a.db.getDumpByPublicID(publicID)
	if err != nil {
		if err == sql.ErrNoRows {
			notFound(w)
			return
		}

		log.Printf("getting file from database error: %v\n", err)
		internalServerError(w)
		return
	}

	// The file has been deleted, which means not found is an approperiate
	// error.
	if dump.deletedAt != nil {
		notFound(w)
		return
	}

	// If we've got a bau and bap from the database we should treat the
	// file as protected and apply basic auth.
	isProtected := false
	var username []byte
	var password []byte
	if len(*dump.username) > 0 && len(*dump.password) > 0 {
		isProtected = true

		if username, err = a.decrypt(*dump.username); err != nil {
			log.Printf("error when decrypting username, %v\n", err)
			internalServerError(w)
			return
		}
		if password, err = a.decrypt(*dump.password); err != nil {
			log.Printf("error when decrypting username, %v\n", err)
			internalServerError(w)
			return
		}
	}

	// Fetch basic auth credentials from the request. If we've got a
	// protected resource but no basic auth credentials were found we'll
	// return a 401.
	u, p, isBasicAuth := r.BasicAuth()
	if isProtected && !isBasicAuth {
		unauthorized(w)
		return
	}

	// The resource is protected and we've got credentials that we can
	// compare, if the credentials doesn't match we'll return a 401.
	if isProtected && isBasicAuth && (subtle.ConstantTimeCompare([]byte(u), username) != 1 ||
		subtle.ConstantTimeCompare([]byte(p), password) != 1) {
		unauthorized(w)
		return
	}

	// Read the file from the filesystem.
	data, err := ioutil.ReadFile(filepath.Join(a.dataDir, dump.filesystemID))
	if err != nil {
		log.Printf("unable to find %s in %s\n", dump.filesystemID, a.dataDir)
		notFound(w)
		return
	}

	// Insert an entry to the access log.
	if err = a.db.insertDumpAccessLog(&dumpAccessLog{
		dumpID:    dump.id,
		ipAddress: r.RemoteAddr,
	}); err != nil {
		log.Printf("unable to insert access log: %v\n", err)
	}

	// Serve the requested file.
	w.Header().Set("Content-Type", dump.contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
