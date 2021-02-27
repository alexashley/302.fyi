MAKEFLAGS += --silent
.PHONY: image run-image run fmt test
default:
	echo No default rule

image:
	docker build -t ghcr.io/alexashley/302.fyi:latest .

run-image: image
	docker run \
		-it \
		--rm \
		-p 1234:1234 \
		ghcr.io/alexashley/302.fyi:latest

run:
	go run main.go

fmt:
	gofmt -s -w .

test:
	go test -v
