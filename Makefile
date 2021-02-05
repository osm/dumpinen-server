PHONY: all
all:
	go build ${LDFLAGS}

armv6:
	CC=arm-linux-gnueabi-gcc GOARCH=arm GOARM=6 go build

install:
	go install

clean:
	rm dumpinen
