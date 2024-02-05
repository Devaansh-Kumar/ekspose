FROM golang:1.21.6-alpine

WORKDIR /controller

COPY ./go.mod /controller/

RUN go mod download

COPY . /controller/

RUN go build -o ekspose

ENTRYPOINT [ "/controller/ekspose" ]