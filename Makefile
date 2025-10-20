# Define variables
BINARY_NAME=bin/mshkbot
LDFLAGS=

# Download dependencies
.PHONY: deps
deps:
	go mod download

# Build binary
.PHONY: build
build: deps
	CGO_ENABLED=0 GOOS=linux go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/mshkbot

# Build Docker image
.PHONY: docker-build
docker-build: build
	docker buildx build -t sukalov/mshkbot --platform linux/amd64 .

# Push Docker image
.PHONY: docker-push
docker-push:
	docker push sukalov/mshkbot:latest

# Development run with Air hot reload
.PHONY: dev
dev:
	air

# Clean up old Docker images
.PHONY: docker-clean
docker-clean:
	ssh root@${DEPLOY_HOST} "\
		docker stop mshk || true; \
		docker rm mshk || true; \
		docker image prune -f \
	"

# Deployment command
.PHONY: deploy
deploy: build docker-build docker-push docker-clean
	ssh root@${DEPLOY_HOST} "\
		docker pull sukalov/mshkbot:latest; \
		docker run --name mshk \
		--restart always \
		--env-file ./mshk/.env -v \
		$(pwd)/root/.env:/root/.env \
		-d sukalov/mshkbot:latest \
	"