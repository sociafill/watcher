.PHONY: all build test gitlab

all: build

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

test:
	go test

gitlab:
	chmod +x .gitlab/init.sh && .gitlab/init.sh