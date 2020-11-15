# Makefile 

.DEFAULT_GOAL := all

VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
            echo v0)

build: 
	go build -o virtualpaper .

release: 
	go build \
		-tags release \
		-ldflags '-X tryffel.net/go/virtualpaper/config.Version=$(VERSION)' \
		-o virtualpaper .

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


