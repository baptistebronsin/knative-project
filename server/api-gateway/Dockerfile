FROM golang:1.24-alpine

WORKDIR /app

COPY go.* .

RUN go mod tidy && go mod verify && go mod download

COPY . .

# docker build -t registry.gitlab.com/baptiste.bronsin/knative/api-gateway:0.0.7 .
# docker push registry.gitlab.com/baptiste.bronsin/knative/api-gateway:0.0.7