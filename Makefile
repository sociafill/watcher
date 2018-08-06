.PHONY: build dockerize run

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o watcher .

dockerize:
	docker build . -t watcher-dev

run:
	docker run -it -p 8000:8000 watcher-dev