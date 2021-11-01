# This makefile creates service images and pushes them to ECR

registry = <aws_account_id>.dkr.ecr.<region>.amazonaws.com # insert your registry
version ?= latest

ecr: build tag push

build:
	docker build -t microservices/server:${version} -f server.Dockerfile .
	docker build -t microservices/cache:${version} -f cache.Dockerfile .
	docker build -t microservices/database:${version} -f database.Dockerfile .

tag:
	docker tag microservices/server:latest ${registry}/microservices/server:${version}
	docker tag microservices/cache:latest ${registry}/microservices/cache:${version}
	docker tag microservices/database:latest ${registry}/microservices/database:${version}

push:
	docker push ${registry}/microservices/server:${version}
	docker push ${registry}/microservices/cache:${version}
	docker push ${registry}/microservices/database:${version}
