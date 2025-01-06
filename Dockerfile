FROM golang:bullseye

RUN apt-get update && apt-get install python3-pip -y && \
  pip install bs4

COPY go.mod go.sum /app/
WORKDIR /app/
RUN go mod download
COPY . /app/
RUN go build -o /usr/bin/monitor2 /app/

ENTRYPOINT ["/usr/bin/monitor2"]
