
.PHONY: postgres
compose-up:
	podman-compose ps
	podman-compose down
	podman-compose up --build

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