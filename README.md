# dicom-go-server

This is a POC Dicom processor written in Go. It has three endpoints:

- `/upload`
- `/header`
- `/image`

The concept is to take a DICOM file on your machine and upload it to the `/upload` endpoint. The application returns a unique file name which you can then query the DICOM tags for (`/header`) or view the DICOM image as a png (`/image`). Commands can be run via curl, Postman or Browser

## Diagram

## Requirements for local usage/setup

- Go
- Docker
- Some downloaded DICOM files on your machine

## Notes

- There are _two_ docker images you can run:

1. [docker-go-server](./docker_image/docker-go-server/)
2. [docker-go-server-with-frontend](./docker_image/dicom-go-server-with-frontend/)

- Each image can be built and run locally using docker
- The frontend is plain bootstrap, js and html ("...tell me you're not a frontend developer, without telling me you're not a frontend developer)
- In the real world, the frontend service and backend api service would be de-coupled to avoid having to deploy both at the same time, but given this is a POC, I've gone a simplified route
- Building a docker image and pushing to `fly.io` made the most sense for the POC work. In the real world, we could run this completely serverless on AWS, GCP or Azure
- Please see the [Examples](#examples) section for how to use the service

## Examples

### Live Endpoints

- With Frontend: https://dicom-go-server.fly.dev/
- API only: https://dicom-go-server-api-only.fly.dev/

### Example Commands

For API only endpoint:

```
curl -F "dicom=@IM000003" https://dicom-go-server-api-only.fly.dev/upload

# This will return your unqiue filename e.g.
{"file":"1746810673989456460_IM000003"}

# Use that unique file name to query the headers e.g.
curl https://dicom-go-server-api-only.fly.dev/header\?file\=1746810673989456460_IM000003\&tag\=00080080

# Use that unqiue file name to view the image in a browser e.g.
open https://dicom-go-server-api-only.fly.dev/image?file=1746810673989456460_IM000003
```

### Example Usage in Browser

https://github.com/user-attachments/assets/da43fdec-4e55-44b1-b115-794a212adabd


## Example DICOM files

See [DICOM_Files](./DICOM_Files/)
