# dicom-go-server-with-frontend docker image

## How to build

`docker build -t dicom-go-server-with-frontend .`

## How to run locally

`docker run -p 8080:8080 -v $(pwd)/uploads:/app/uploads dicom-go-server-with-frontend`

## How to test

Go to http://locahost:8080/ on your browser and follow the steps!
