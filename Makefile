tests:
	go test ./... -cover 

start-with-migrations:
	docker-compose -f ./deployments/docker-compose.migrate.yaml up --build	

start:
	docker-compose -f ./deployments/docker-compose.yaml up --build	

stop:
	docker container stop $$(docker container list -q)

build:
	docker-compose  -f ./deployments/docker-compose.yaml build