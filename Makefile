# Makefile 

.DEFAULT_GOAL := all

VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
            echo v0)
COMMIT ?= $(shell git describe --always --dirty 2> /dev/null || \
            echo v0)

build: set-commit
	go build -o virtualpaper .

release: set-commit
	go build \
		-tags release \
		-o virtualpaper .

set-commit:
	sed -i 's/var Commit = ".*"/var Commit = "$(COMMIT)"/g' config/version.go

run: 
	go run main.go

test: 
	go test ./...

run-frontend: 
	cd frontend; yarn start

build-frontend: 
	cd frontend; REACT_APP_STAGE=prod yarn build


swagger:
	swagger serve -F=swagger swagger.yaml

build-swagger:
	swagger generate spec -o ./swagger.yaml --scan-models
	

all: test release build-frontend 


