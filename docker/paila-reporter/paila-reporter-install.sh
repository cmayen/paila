#!/bin/bash
#
# This script calls docker compose up and detaches it. 
#
# Author: Chris Mayenschein
# GitHub: https://github.com/cmayen/paila
# Date: 2025-07-21
# Last Modified: 2025-07-22
#
# Usage: ./paila-reporter-install.sh
#
################################################################################

# make sure the shared volume exists
sudo docker volume create --name paila_ingest_data

# Bring up the services via docker-compose
sudo docker compose up -d

