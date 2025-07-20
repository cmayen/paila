#!/bin/bash
#
# This script takes a fresh install of ubuntu server minimized and updates
# the system while installing docker and other necessary packages to run
# ollama in a container with amd support.
#
# Author: Chris Mayenschein
# GitHub: https://github.com/cmayen/paila
# Date: 2025-07-20
# Last Modified: 2025-07-20
#
# Usage: ./install_docker_ubuntu.sh
#
################################################################################


# Do your updates!
sudo apt update -y && sudo apt upgrade -y


# Retrieve the docker and jq packages. It would be a good idea to have an 
# editor like vim, also ensure wget is installed, and gpg will be needed 
# for keys. One call sounds good for all of them.
sudo apt install -y docker.io jq vim wget gpg


# Get the AMD GPU installer package from the radeon repo and install 
# it. (At the time of this document, the version was 6.3.x)
wget https://repo.radeon.com/amdgpu-install/6.3.4/ubuntu/jammy/amdgpu-install_6.3.60304-1_all.deb
sudo apt install -y ./amdgpu-install_6.3.60304-1_all.deb
sudo amdgpu-install --usecase=dkms

 
# The amd-container-toolkit is required to get everything working so the 
# containers can talk to the GPU. The package is not available by the default 
# repositories, so we need to add the information.


# First get the keys:
wget https://repo.radeon.com/rocm/rocm.gpg.key -O - | gpg --dearmor | sudo tee /etc/apt/keyrings/rocm.gpg > /dev/null


# Now add the repo information to the sources:
echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amd-container-toolkit/apt/ jammy main" | sudo tee /etc/apt/sources.list.d/amd-container-toolkit.list


# In order for apt to know the repo exists, we need to update it. We may 
# as well check for updates/upgrades too.
sudo apt update -y && sudo apt upgrade -y


# Install the amd-container-toolkit and docker-compose-v2
sudo apt install amd-container-toolkit docker-compose-v2


# Configure the runtime, and restart docker.
sudo amd-ctk runtime configure
sudo systemctl restart docker


# Test that the GPU is available using rocm/rocm-terminal
#sudo docker run --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/rocm-terminal amd-smi monitor

# You should see GPUs listed, in my case, there is only one.
# GPU  POWER   GPU_T   MEM_T   GFX_CLK   GFX%   MEM%   ENC%   DEC%      VRAM_USAGE
#   0   10 W   47 °C   46 °C   800 MHz    0 %    0 %    N/A    N/A    0.0/  8.0 GB


#Looks good, let’s get a list of the docker containers so we can remove 
# the one we just ran. Good housekeeping and all.
sudo docker ps -a
#CONTAINER ID   IMAGE                COMMAND               CREATED          STATUS                      PORTS                                           NAMES
#30c6e6bfbfe2   rocm/rocm-terminal   "amd-smi monitor"     2 minutes ago   Exited (0) 2 minutes ago                                                   boring_cartwright


# Remove the test container
#sudo docker remove 30c6e6bfbfe2



