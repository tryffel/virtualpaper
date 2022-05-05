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

test-unit: 
	# skip e2e tests
	go test ./... -short

test-e2e:
	go test ./e2e

test-start: export TEST_VOLUME_ID=-test
test-start:
	docker-compose up -d
	# wait for db to start
	sleep 3
	docker-compose run --rm --entrypoint "/app/virtualpaper --config /config/config.toml" server migrate
	docker-compose run --rm --entrypoint "/app/virtualpaper --config /config/config.toml" server manage add-user -U user -P user -a false
	docker-compose start server


#test-start:
#	export TEST_VOLUME_ID=test
#	docker-compose up -d 
#

test-stop: export TEST_VOLUME_ID=-test
test-stop:
	docker-compose down -v

test-cleanup: export TEST_VOLUME_ID=-test
test-cleanup:
	docker volume rm virtualpaper_data
	docker volume rm virtualpaper_postgres
	docker volume rm virtualpaper_meilisearch

test: test-start test-e2e test-stop


run-frontend: 
	cd frontend; yarn start

build-frontend: 
	cd frontend; REACT_APP_STAGE=prod yarn build


swagger:
	swagger serve -F=swagger swagger.yaml

build-swagger:
	swagger generate spec -o ./swagger.yaml --scan-models
	

all: test release build-frontend 


