.PHONY: all server server_test ng clean

all:
	make server
	make server_test
	make ng

server:
	go build -o server goserver/server.go

server_test:
	go build -o server_test goserver/testing/test.go

ng:
	ng build --prod --aot

clean:
	rm -f server
	rm -rf dist
