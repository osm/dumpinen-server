# dumpinen

Dumpinen is a tiny storage service

An uploaded file will be deleted if the deleteAfter query parameter is set to
a valid duration, the file will be stored forever if the parameter is omitted.

It is also possible to password protect the resource by setting a basic auth
username and password when the file is uploaded.

Check out cli at https://github.com/osm/dumpinen

## Requirements

* A PostgreSQL server
* A directory to store the files in
* A set of public and private keys from age, see https://github.com/FiloSottile/age

## Start a server

```sh
$ go build
$ age-keygen >agekey.txt
$ ./dumpinen-server \
	-cs <connection string> \
	-data-dir /tmp \
	-port 8080 \
	-pub-key $(grep "public key:" agekey.txt | awk '{ print $NF }') \
	-priv-key $(tail -n1 agekey.txt)
```

## Upload examples

### Upload a file without expiration time and protection.

```sh
$ echo "foo" >/tmp/foo.txt
$ curl --data-binary @/tmp/foo.txt http://localhost:8080
GAKJObQturg
$ curl http://localhost:8080/GAKJObQturg
foo
```

### Upload a file which will be deleted when the duration has passed.

See https://golang.org/pkg/time/#ParseDuration for more information about the
duration format.

```sh
$ echo "foo" >/tmp/foo.txt
$ curl --data-binary @/tmp/foo.txt http://localhost:8080?deleteAfter=5s
n5-IluF9tsq
$ curl http://localhost:8080/n5-IluF9tsq
foo
$ sleep 5s
$ curl http://localhost:8080/n5-IluF9tsq
not found
```

### Upload a file and protect it with basic auth.

```sh
$ echo "foo" >/tmp/foo.txt
$ curl -u foo:bar --data-binary @/tmp/foo.txt http://localhost:8080
N64agNx9woL
$ curl http://localhost:8080/N64agNx9woL
unauthorized
$ curl -u foo:bar http://localhost:8080/N64agNx9woL
foo
```

### Upload a file with a custom content type


```sh
$ echo "foo" >/tmp/foo.txt
$ curl --data-binary @/tmp/foo.txt http://localhost:8080?contentType=plain/text
n5-IluF9tsq
$ curl http://localhost:8080/n5-IluF9tsq
foo
```

## Routes

| Method | Route  | Query parameters                              |
| ------ | ------ | --------------------------------------------- |
| POST   | /      | deleteAfter=duration, contentType=contentType |
| GET    | /:id   |                                               |
