FROM golang:1.18
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o raft-api cmd/raft-api/main.go
CMD ["/app/raft-api"]