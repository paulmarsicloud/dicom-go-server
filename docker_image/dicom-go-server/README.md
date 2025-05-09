# dicom-go-server docker image

## How to build

`docker build -t dicom-go-server .`

## How to run locally

`docker run -p 8080:8080 -v $(pwd)/uploads:/app/uploads dicom-go-server`

## How to test

```
curl -F "dicom=@IM000003" https://localhost:8080/upload

# This will return your unqiue filename e.g.
{"file":"1746810673989456460_IM000003"}

# Use that unique file name to query the headers e.g.
curl https://localhost:8080/header\?file\=1746810673989456460_IM000003\&tag\=00080080

# Use that unqiue file name to view the image in a browser e.g.
open https://localhost:8080/image?file=1746810673989456460_IM000003
```
