.PHONY: build run test lint clean

build:
	CGO_ENABLED=1 go build -o peaktop ./cmd/peaktop/

run:
	CGO_ENABLED=1 go run ./cmd/peaktop/

test:
	CGO_ENABLED=1 go test ./internal/...

lint:
	go vet ./...
	go fmt ./...

clean:
	rm -f peaktop
