.PHONY: build
build:
	go build .

.PHONY: build-docker-builder
build-docker-builder:
	@docker build \
       --target builder \
       --cache-from aureolecloud/aureole-builder:latest \
       -t aureolecloud/aureole-builder:latest \
       -f Dockerfile .

.PHONY: build-docker-image
build-docker-image:
	@docker build \
      --cache-from aureolecloud/aureole-builder:latest \
      --cache-from aureolecloud/aureole:latest \
      -t aureolecloud/aureole:latest \
      -f Dockerfile .