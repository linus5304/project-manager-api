.PHONY: test fmt fmt-check tidy run

test:
	go test ./...

fmt:
	gofmt -w .

fmt-check:
	test -z "$$(gofmt -l .)"

tidy:
	go mod tidy

run:
	go run ./cmd/api
