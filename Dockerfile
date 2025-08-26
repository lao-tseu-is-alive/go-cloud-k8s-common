# Start from the latest golang base image
FROM golang:1.24.5-alpine3.22 AS builder

# Define build arguments for version and build timestamp
ARG APP_REVISION
ARG BUILD
ARG APP_REPOSITORY=https://github.com/lao-tseu-is-alive/go-cloud-k8s-common

# Add Maintainer Info
LABEL maintainer="cgil"
LABEL org.opencontainers.image.title="go-cloud-k8s-common"
LABEL org.opencontainers.image.description="This is a go-cloud-k8s-common container image, demo for my common packages i use in my Go programs, without standard go http package (not using framework)"
LABEL org.opencontainers.image.url="https://ghcr.io/lao-tseu-is-alive/go-cloud-k8s-common:latest"
LABEL org.opencontainers.image.authors="cgil"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.version="1.0.0"
# Set image version label dynamically
LABEL org.opencontainers.image.source="${APP_REPOSITORY}"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY "cmd/fullServer" ./server
COPY pkg ./pkg

# Clean the APP_REPOSITORY for ldflags
RUN APP_REPOSITORY_CLEAN=$(echo $APP_REPOSITORY | sed 's|https://||') && \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s -X ${APP_REPOSITORY_CLEAN}/pkg/version.REVISION=${APP_REVISION} -X ${APP_REPOSITORY_CLEAN}/pkg/version.BuildStamp=${BUILD}" -o go-common-server ./server


######## Start a new stage  #######
# using from scratch for size and security reason
# Containers Are Not VMs! Which Base Container (Docker) Images Should We Use?
# https://blog.baeke.info/2021/03/28/distroless-or-scratch-for-go-apps/
# https://github.com/vfarcic/base-container-images-demo
# https://youtu.be/82ZCJw9poxM
FROM scratch
# to comply with security best practices
# Running containers with 'root' user can lead to a container escape situation (the default with Docker...).
# It is a best practice to run containers as non-root users
# https://docs.docker.com/develop/develop-images/dockerfile_best-practices/
# https://docs.docker.com/engine/reference/builder/#user
USER 12121:12121
WORKDIR /goapp
COPY certificates/isrg-root-x1-cross-signed.pem /goapp/certificates/
# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/go-common-server .

# Expose port 9999 to the outside world, go-common-server will use the env PORT as listening port or 9999 as default
EXPOSE 9999

# Health check to ensure the app is running
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:9999/health || exit 1

# Command to run the executable
CMD ["./go-common-server"]
