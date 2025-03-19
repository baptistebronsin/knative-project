FROM golang:1.24-alpine

WORKDIR /app

COPY go.* .

RUN go mod tidy && go mod verify && go mod download

COPY . .

# docker build -t registry.gitlab.com/baptiste.bronsin/knative/comments:0.0.6 .
# docker push registry.gitlab.com/baptiste.bronsin/knative/comments:0.0.6