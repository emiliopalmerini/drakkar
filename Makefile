.PHONY: build test fmt vet clean run

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test: vet
	go test ./...

build: vet
	go build -o drakkar .

run: build
	./drakkar

clean:
	rm -f drakkar
	go clean
