APP_NAME=ether
REGISTRY?=343660461351.dkr.ecr.us-east-2.amazonaws.com
TAG?=latest
ETHER_DB_LOCATION?=localhost:3306
ETHER_DB_USERNAME?=ether
ETHER_DB_PASSWORD?=ether
ETHER_CONTENT_DIR?=./
HELP_FUNC = \
    %help; \
    while(<>) { \
        if(/^([a-z0-9_-]+):.*\#\#(?:@(\w+))?\s(.*)$$/) { \
            push(@{$$help{$$2 // 'targets'}}, [$$1, $$3]); \
        } \
    }; \
    print "usage: make [target]\n\n"; \
    for ( sort keys %help ) { \
        print "$$_:\n"; \
        printf("  %-20s %s\n", $$_->[0], $$_->[1]) for @{$$help{$$_}}; \
        print "\n"; \
    }

.PHONY: help
help: 				## show options and their descriptions
	@perl -e '$(HELP_FUNC)' $(MAKEFILE_LIST)

all:				## clean the working environment, build and test the packages, and then build the docker image
all: clean test docker

tmp:				## create tmp/
	if [ -d "./tmp" ]; then rm -rf ./tmp; fi
	mkdir tmp

build: tmp			## build the app binaries
	go build -o ./tmp ./...

test: build 		## build and test the module packages
	export ETHER_CONTENT_DIR=${ETHER_CONTENT_DIR} && \
		go test ./...

run: build			## build and run the app binaries
	export ETHER_DB_LOCATION=${ETHER_DB_LOCATION} && \
		export ETHER_DB_USERNAME=${ETHER_DB_USERNAME} && \
		export ETHER_DB_PASSWORD=${ETHER_DB_PASSWORD} && \
		export ETHER_CONTENT_DIR=${ETHER_CONTENT_DIR} && \
		./tmp/app

docker: tmp 		## build the docker image
	docker build -t $(REGISTRY)/$(APP_NAME):$(TAG) .

docker-run: docker	## start the built docker image in a container
	docker run -p 80:80 \
		-e ETHER_DB_LOCATION=$(ETHER_DB_LOCATION) \
		-e ETHER_DB_USERNAME=$(ETHER_DB_USERNAME) \
		-e ETHER_DB_PASSWORD=$(ETHER_DB_PASSWORD) \
		-e ETHER_CONTENT_DIR=$(ETHER_CONTENT_DIR) \
		--name $(APP_NAME) $(REGISTRY)/$(APP_NAME):$(TAG)

docker-push: tmp docker
	docker push $(REGISTRY)/$(APP_NAME):$(TAG)

.PHONY: clean
clean:				## remove tmp/, stop and remove app container, old docker images
	rm -rf tmp
ifneq ("$(shell docker container list -a | grep $(APP_NAME))", "")
	docker rm -f $(APP_NAME)
endif
	docker system prune
ifneq ("$(shell docker images | grep $(APP_NAME) | awk '{ print $$3; }')", "")
	docker images | grep $(APP_NAME) | awk '{ print $$3; }' | xargs -I {} docker rmi -f {}
endif
