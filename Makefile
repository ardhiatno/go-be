IMAGE_NAME := go-be
IMAGE_TAG  := latest
CONTAINER_NAME := go-be

# Build docker image
build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .