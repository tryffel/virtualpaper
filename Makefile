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
	go test ./api ./cmd ./config ./errors ./mail ./models ./process ./search ./storage

test-external:
	# test integration with external programs
	go test ./... -short -tags test_integration

test-api:
	# diseble caching
	go test -v ./integration_tests -count=1


test-clear-records:
	docker exec -it virtualpaper_postgres_1 psql -U virtualpaper -c \
		"delete from documents where 1=1; delete from metadata_values where 1=1; delete from metadata_keys where 1=1;"


test-e2e:
	cd frontend; node_modules/.bin/cypress run

test-start: export TEST_VOLUME_ID=-test
test-start:
	docker-compose up -d
	# wait for db to start
	sleep 3
	docker-compose run --rm --entrypoint "/app/virtualpaper --config /config/config.toml" server migrate
	docker-compose run --rm --entrypoint "/app/virtualpaper --config /config/config.toml" server manage add-user -U user -P superstronguser -a false
	docker-compose run --rm --entrypoint "/app/virtualpaper --config /config/config.toml" server manage add-user -U admin -P superstrongadmin -a true
	docker-compose start server

ci-test-unit:
	 /go/bin/gotestsum --format testname ./api ./cmd ./config ./errors ./mail ./models ./process ./search ./storage

ci-test-api:
	/go/bin/gotestsum --format testname ./integration_tests -count=1

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
	rm integration_tests/TOKEN

test: test-unit test-integration test-start test-api test-e2e test-stop


run-frontend: 
	cd frontend; yarn start

build-frontend: 
	cd frontend; REACT_APP_STAGE=prod yarn build


swagger:
	swagger serve -F=swagger swagger.yaml

build-swagger:
	swagger generate spec -o ./swagger.yaml --scan-models


dev-init:
	docker volume create virtualpaper-dev-go
	mkdir -p dev/config dev/logs dev/data
	cp config.sample.toml dev/config/config.toml
	echo "Please edit the config file in dev/config/config.toml. See docker-compose.yml for help."

dev-build-container: 
	docker build --file=Dockerfile.dev -t tryffel/virtualpaper-devenv:latest .
	docker volume create virtualpaper-dev-go

dev-debug:
	echo "Starting debug docker container"
	docker run --name=virtualpaper-dev \
		--rm -d -it \
		-p 127.0.0.1:8000:8000 \
		-p 127.0.0.1:2345:2345 \
		-v `pwd`:/virtualpaper \
		-v `pwd`/dev/config:/config \
		-v `pwd`/dev/data:/data \
		-v `pwd`/dev/logs:/logs \
		--network virtualpaper_virtualpaper \
		-v virtualpaper-dev-go:/go/pkg/ \
		tryffel/virtualpaper-devenv:latest /bin/sh
	echo "Starting dlv inside container"
	docker exec -it virtualpaper-dev \
		/bin/sh -c "dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient  -- serve --config /config/config.toml"
	docker kill virtualpaper-dev

dev-start-container:
	echo "Starting docker container"
	docker run --name=virtualpaper-dev \
		--rm -d -it \
		-p 127.0.0.1:8000:8000 \
		-p 127.0.0.1:2345:2345 \
		-v `pwd`:/virtualpaper \
		-v `pwd`/dev/config:/config \
		-v `pwd`/dev/data:/data \
		-v `pwd`/dev/logs:/logs \
		--network virtualpaper_virtualpaper \
		-v virtualpaper-dev-go:/go/pkg/ \
		tryffel/virtualpaper-devenv:latest /bin/sh
	docker exec -it virtualpaper-dev \
		/bin/sh -c "go run . serve --config /config/config.toml"
	docker kill virtualpaper-dev

dev-start-container-no-server:
	echo "Starting docker container"
	docker run --name=virtualpaper-dev \
		--rm -it \
		-v `pwd`:/virtualpaper \
		-v `pwd`/dev/config:/config \
		-v `pwd`/dev/data:/data \
		-v `pwd`/dev/logs:/logs \
		--network virtualpaper_virtualpaper \
		-v virtualpaper-dev-go:/go/pkg/ \
		tryffel/virtualpaper-devenv:latest /bin/sh






all: test release build-frontend 


