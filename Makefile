all: clean build

clean:
	rm -vf ./bin/issues-server

build:
	mkdir -p ./bin
	go build -o ./bin/issues-server cmd/issues-server/main.go
