APP_NAME=ether
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
	go test ./...

run: build			## build and run the app binaries
	./tmp/app

docker: tmp 		## build the docker image
	docker build -t $(APP_NAME) .

docker-run: docker	## start the built docker image in a container
	docker run -p 80:8080 --env-file ./env.list --name $(APP_NAME) $(APP_NAME)

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
