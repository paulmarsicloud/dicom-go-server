FROM golang:1.21-alpine

# 1) Install git (needed to fetch modules)
RUN apk add --no-cache git

WORKDIR /app

# 2) Copy only your main.go in
COPY main.go ./

# 3) Initialize a module, pull in the DICOM library, and tidy
RUN go mod init dicom-server \
    && go get github.com/suyashkumar/dicom@v1.0.4 \
    && go mod tidy

# 4) Build your binary
RUN go build -o server

EXPOSE 8080
CMD ["./server"]
