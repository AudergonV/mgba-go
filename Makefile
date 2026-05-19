.PHONY: build test clean demo

build:
	go build -v ./...

test:
	go test -v ./pkg/mgba/

demo:
	go run ./cmd/demo/main.go $(ROM)

clean:
	go clean
	rm -f demo