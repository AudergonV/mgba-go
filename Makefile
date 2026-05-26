.PHONY: build test test-race test-bench clean demo

build:
	go build -v ./...

test:
	go test -v -count=1 ./pkg/mgba/

test-race:
	go test -v -count=1 -race ./pkg/mgba/

test-bench:
	go test -v -count=1 -bench=. -benchmem ./pkg/mgba/

build-demo:
	go build -o demo ./cmd/demo/

demo:
	go run ./cmd/demo/main.go $(ROM)

clean:
	go clean
	rm -f demo