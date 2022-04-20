# The base go-image
FROM golang:1.18-alpine

# Install git to download dependencies
#RUN apk update
#RUN apk add git

# Create a directory for the app
RUN mkdir /app
RUN mkdir /config.d

# Copy all files from the current directory to the app directory
COPY keycloakUpdater /app

# Set working directory
WORKDIR /app

# Run command as described:
# go build will build an executable file named server in the current directory
#RUN go build -o keycloakUpdater .

# Run the server executable
ENTRYPOINT [ "/app/keycloakUpdater" ]