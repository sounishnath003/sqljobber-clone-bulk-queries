
Appname=gooferr-$$(date +%Y%m%d)

.PHONY: postgres
postgres:
	podman run -ti -p 5432:5432 \
		-e POSTGRES_USER=root \
		-e POSTGRES_PASSWORD=root \
		-e PGDATA=/var/lib/postgresql/data/pgdata \
		-v ./postgres:/var/lib/postgresql/data \
		--name pqdb \
		postgres:15-alpine

.PHONY: createdb
createdb:
	podman exec -ti pqdb createdb --username=root --owner=root gochat
	podman exec -ti psql


.PHONY: build
build:
	go build -o dist/$(Appname) cmd/*.go
	ls -l dist


.PHONY: run
run: build
	dist/$(Appname) --race