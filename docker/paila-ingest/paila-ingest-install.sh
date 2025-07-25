#!/bin/bash
#
# This script calls docker compose up and detaches it. 
#
# Author: Chris Mayenschein
# GitHub: https://github.com/cmayen/paila
# Date: 2025-07-20
# Last Modified: 2025-07-23
#
# Usage: ./paila-ingest-install.sh
#
################################################################################

# make sure the shared volume exists
sudo docker volume create --name paila_ingest_data

# Bring up the services via docker-compose
sudo docker compose up -d

