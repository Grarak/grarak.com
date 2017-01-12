.PHONY: all server ng clean

all:
	make server
	make ng

server:
	go build -o server goserver/server.go

ng:
	ng build --prod --aot

clean:
	rm -f server
	rm -rf dist
