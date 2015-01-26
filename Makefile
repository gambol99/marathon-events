#
#   Author: Rohith (gambol99@gmail.com)
#   Date: 2015-01-26 12:59:54 +0000 (Mon, 26 Jan 2015)
#
#  vim:ts=2:sw=2:et
#
NAME=marathon-events
AUTHOR=gambol99
HARDWARE=$(shell uname -m)
VERSION=$(shell awk '/const Version/ { print $$4 }' version.go | sed 's/"//g')
REPOSITORY=embassy

.PHONY: build docker release

build:
	mkdir -p stage/
	go get
	go build -o stage/${NAME}

docker: build
	docker build -t ${AUTHOR}/${NAME} .

all: clean build docker

clean:
	rm -f ./stage/${NAME}
	rm -rf ./release
	go clean

release:
	rm -rf release
	mkdir -p release
	GOOS=linux godep go build -o release/$(NAME)
	cd release && gzip -c ${NAME} > $(NAME)_$(VERSION)_linux_$(HARDWARE).gz
	GOOS=darwin godep go build -o release/$(NAME)
	cd release && gzip -c ${NAME} > $(NAME)_$(VERSION)_darwin_$(HARDWARE).gz
	rm release/$(NAME)

changelog: release
	git log $(shell git tag | tail -n1)..HEAD --no-merges --format=%B > release/changelog


