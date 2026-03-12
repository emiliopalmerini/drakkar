.PHONY: build test lint fmt vet clean run

build:
	go build -o drakkar .

test:
	go test ./...

lint: vet
	@echo "lint: done (vet only, no external tools)"

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f drakkar
	go clean

run: build
	./drakkar
