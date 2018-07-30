.PHONY: all build test

all: build

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

test:
	go test