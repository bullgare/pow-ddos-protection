
# runs server and client in docker
.PHONY: run-in-docker
.SILENT: run-in-docker
run-in-docker:
	docker-compose -f ./build/dev/docker-compose.yaml up

# force rebuilds and runs server and client in docker
.PHONY: rebuild-and-run-in-docker
.SILENT: rebuild-and-run-in-docker
rebuild-and-run-in-docker:
	docker-compose -f ./build/dev/docker-compose.yaml up --build client server

