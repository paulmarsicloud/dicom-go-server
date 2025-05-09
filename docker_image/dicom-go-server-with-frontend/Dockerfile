FROM golang:1.21-alpine

RUN apk add --no-cache git

WORKDIR /app

COPY main.go ./

RUN go mod init dicom-server \
    && go get github.com/suyashkumar/dicom@v1.0.4 \
    && go mod tidy

RUN go build -o server

EXPOSE 8080
CMD ["./server"]
