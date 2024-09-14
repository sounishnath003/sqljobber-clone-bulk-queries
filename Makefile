
.PHONY: postgres
db:
	podman run --rm -d -ti \
		-e POSTGRES_USER=root \
		-e POSTGRES_PASSWORD=root \
		-p 5432:5432 \
		postgres:16-alpine

.PHONY: install
install:
	go mod tidy
	go mod download
	go mod verify

.PHONY: build
build:
	go build -ldflags "-s -w" -o dist/gooferr cmd/*.go
	ls -lrth dist

.PHONY: run
run: build
	./dist/gooferr