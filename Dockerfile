############################
# STEP 1 build executable binary
############################
FROM golang:1.22.4-alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

RUN mkdir /build
ADD . /build/
WORKDIR /build

# Fetch dependencies.
RUN go get -d -v

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o keycloakUpdater .

############################
# STEP 2 build a small image
############################
FROM scratch
COPY --from=builder /build/keycloakUpdater /app/
WORKDIR /workspace
ENTRYPOINT ["/app/keycloakUpdater"]
