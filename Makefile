.PHONY: all build-api build-ui docker-api docker-ui clean

all: build-api build-ui docker-api docker-ui

build-api:
	@go build -o rahanna-api cmd/api/main.go

build-ui:
	@go build -o rahanna-ui cmd/ui/main.go

docker-api:
	@docker build -t rahanna-api:latest -f docker/api/Dockerfile .

docker-ui:
	@docker build -t rahanna-ui:latest -f docker/ui/Dockerfile .

clean:
	-@docker rmi -f rahanna-api rahanna-ui
