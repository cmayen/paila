#!/bin/bash
#
# This script creates a docker image for paila-reporter automatically based
# on the ubuntu:latest image, updating it and adding/installing packages
# and files onto the filesystem. 
# The docker image is saved to the local repo as paila-reporter:latest
# The image is backed up via docker save to paila-reporter-latest.tar.gz
#
# Author: Chris Mayenschein
# GitHub: https://github.com/cmayen/paila
# Date: 2025-07-21
# Last Modified: 2025-07-23
#
# Usage: ./paila-reporter-image-create.sh
#
################################################################################


# check for existing container
piiExists=$(sudo docker ps | grep -c -m 1 paila-reporter-image)
if [[ "$piiExists" == 1 ]]; then
  echo ""
else
  # running image container does not exist yet, create one

  # create the base image paila-reporter will be installed on
  #sudo docker run -dit -p 80:80 --name paila-reporter ubuntu:latest
  d_pi=$(sudo docker run -dit --name paila-reporter-image ubuntu:latest)
  # returns: 1d7fad2f50785ab43bca8eec2997190ab12e4a8771e7860095a668aef4255807

  # run updates on the new container
  sudo docker exec -it paila-reporter-image apt-get update
  sudo docker exec -it paila-reporter-image apt-get upgrade -y

fi


# create necessary directories, (tied into the ingest storage)
sudo docker exec -it paila-reporter-image mkdir .paila-ingest
sudo docker exec -it paila-reporter-image mkdir .paila-ingest/uploads
sudo docker exec -it paila-reporter-image mkdir .paila-ingest/reports
sudo docker exec -it paila-reporter-image mkdir .paila-ingest/archive


# copy reporter server go binary and other files into place
sudo docker cp paila-reporter-go paila-reporter-image:/
sudo docker exec -it paila-reporter-image mkdir .paila-reporter
sudo docker exec -it paila-reporter-image mkdir .paila-reporter/public

#
for f in public/*; do sudo docker cp $f paila-reporter-image:.paila-reporter/public/; done


# test server
#sudo docker exec -it paila-reporter-image ./paila-reporter-go


# commit the new image locally
sudo docker commit paila-reporter-image paila-reporter:latest
# responds:sha256:12ea724adc2d6d1487378daf2335606653ec468683dfd0208cfa1ca7406ca087


# save the image to a file
sudo docker image save paila-reporter:latest | gzip > paila-reporter-latest.tar.gz


# load the image on a new machine
#








# commit the new image to dockerhub







# load the image into kub for pods














