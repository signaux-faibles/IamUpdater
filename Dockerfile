FROM golang:1.18-alpine as builder
RUN apk add --no-cache git
RUN mkdir /build
# must create 'workspace' folder here because scratch image can't create it
RUN mkdir /workspace
ADD . /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o keycloakUpdater .

FROM scratch
COPY --from=builder /build/keycloakUpdater /app/
COPY --from=builder /workspace /workspace
WORKDIR /app
CMD ["./keycloakUpdater"]