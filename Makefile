
# runs server and client in docker
.PHONY: run-in-docker
.SILENT: run-in-docker
run-in-docker:
	docker-compose -f ./build/dev/docker-compose.yaml up --force-recreate -V \
		&& docker-compose -f ./build/dev/docker-compose.yaml rm -fsv

# force rebuilds and runs server and client in docker
.PHONY: rebuild-and-run-in-docker
.SILENT: rebuild-and-run-in-docker
rebuild-and-run-in-docker:
	docker-compose -f ./build/dev/docker-compose.yaml up --force-recreate -V --build server client1 client2 \
		&& docker-compose -f ./build/dev/docker-compose.yaml rm -fsv


# run unit tests
.PHONY: test
.SILENT: test
test:
	go test -v -short -race ./...

# run integration tests
.PHONY: test-integration
.SILENT: test-integration
test-integration:
	go test -test.run Integration --tags=integration ./...

