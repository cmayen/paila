#!/bin/bash
#
# This script takes a fresh install of ubuntu server minimized and updates
# the system while installing docker and other necessary packages to run
# ollama in a container with amd support.
#
# Author: Chris Mayenschein
# GitHub: https://github.com/cmayen/paila
# Date: 2025-07-21
# Last Modified: 2025-07-21
#
# Usage: ./install-docker-nvidia-ubuntu.sh
#
################################################################################


# Do your updates!
sudo apt update -y && sudo apt upgrade -y


# Retrieve the docker and jq packages. It would be a good idea to have an 
# editor like vim, also ensure wget is installed, and gpg will be needed 
# for keys. Also need docker-compose-v2 One call sounds good for all of them.
sudo apt install -y docker.io jq vim wget gpg docker-compose-v2


#
# https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html
#

# configure apt for the nvidia production repository
curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey | sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg \
  && curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list | \
    sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' | \
    sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list


# Optionally, configure the repository to use experimental packages:
#sed -i -e '/experimental/ s/^#//g' /etc/apt/sources.list.d/nvidia-container-toolkit.list


# In order for apt to know the repo exists, we need to update it. We may 
# as well check for updates/upgrades too.
sudo apt update -y && sudo apt upgrade -y


# Install the nvidia-container-toolkit
# At the time of this document, 1.17.8-1 is latest.
export NVIDIA_CONTAINER_TOOLKIT_VERSION=1.17.8-1
  sudo apt-get install -y \
      nvidia-container-toolkit=${NVIDIA_CONTAINER_TOOLKIT_VERSION} \
      nvidia-container-toolkit-base=${NVIDIA_CONTAINER_TOOLKIT_VERSION} \
      libnvidia-container-tools=${NVIDIA_CONTAINER_TOOLKIT_VERSION} \
      libnvidia-container1=${NVIDIA_CONTAINER_TOOLKIT_VERSION}


# Configure the runtime, and restart docker.
sudo nvidia-ctk runtime configure --runtime=docker
sudo systemctl restart docker

# todo
# need to add test commands here like the amd installer
