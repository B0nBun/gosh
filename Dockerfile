FROM golang:alpine3.18 AS build

# Needed for go-sqlite3
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev

WORKDIR /workspace
COPY . /workspace/

RUN go mod tidy
RUN go install github.com/mattn/go-sqlite3
RUN go build -o /workspace/gosh_server

EXPOSE 1234

CMD [ "/workspace/gosh_server", "-zip", "-addr", "0.0.0.0:1234", "-ds", ":memory:" ]
