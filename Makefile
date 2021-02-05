GO_BUILD_ENV := CGO_ENABLED=0 GOOS=linux GOARCH=amd64
DOCKER_BUILD=$(shell pwd)/.docker_build
DOCKER_CMD=$(DOCKER_BUILD)/PanIndex

$(DOCKER_CMD): clean
	mkdir -p $(DOCKER_BUILD)
	flags="-X 'main.VERSION=master' -X 'main.BUILD_TIME=$(date "+%F %T")' -X 'main.GO_VERSION=$(go version)'-X 'main.GIT_COMMIT_SHA=$(git show -s --format=%H)'"
	$(GO_BUILD_ENV) go build -ldflags="$flags" -v -o $(DOCKER_CMD) .

clean:
	rm -rf $(DOCKER_BUILD)

heroku: $(DOCKER_CMD)
	heroku container:push web