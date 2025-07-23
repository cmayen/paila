#!/bin/bash
sudo docker stop paila-reporter-image
sudo docker rm paila-reporter-image
sudo docker stop paila-reporter
sudo docker rm paila-reporter
go build paila-reporter-go.go
./paila-reporter-image-create.sh
