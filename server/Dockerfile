# Use the official Go image based on your go.mod file
FROM golang:1.23.4-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your application source code
COPY . .

# Build the Go application
RUN go build -o /vivacity-app ./server

# Expose the port your app runs on
EXPOSE 8080

# Command to run the executable
CMD [ "/vivacity-app" ]