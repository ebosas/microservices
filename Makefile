# This makefile creates service images and pushes them to ECR

project = microservices
services = server cache database backend
# Change to your own registry
registry = 123456789012.dkr.ecr.us-west-1.amazonaws.com
version ?= latest

ecr: ecr-build ecr-publish

ecr-build:
	for service in ${services} ; do \
		docker build -t ${project}/$$service:${version} -f $$service.Dockerfile . ; \
	done

ecr-publish:
	for service in ${services} ; do \
		docker tag ${project}/$$service:latest ${registry}/${project}/$$service:${version} ; \
		docker push ${registry}/${project}/$$service:${version} ; \
	done
