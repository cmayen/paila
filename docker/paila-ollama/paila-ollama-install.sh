#!/bin/bash


# Bring up the services via docker-compose
sudo docker compose up -d


# Get the list of docker containers so we can get the id for adding the model.
sudo docker ps -a
#CONTAINER ID   IMAGE                COMMAND               CREATED         STATUS                     PORTS                                           NAMES
#739a8ed9c057   ollama/ollama:rocm   "/bin/ollama serve"   3 minutes ago     Up 3 minutes                 0.0.0.0:11434->11434/tcp, :::11434->11434/tcp   paila-ollama


#Now that we have the container id, we’ll use that to get our model 
# pulled for ollama. It looks like gemma3 might be a good choice for 
# this experiment. Feel free to use and add any models you like that 
# your GPU can handle.
sudo docker exec -it paila-ollama ollama pull gemma3


# Test it! Make a curl request and generate a response with the model 
# chosen, and setting stream to false.
curl http://localhost:11434/api/generate -d '{ "model": "gemma3", "stream": false, "prompt": "Hi Gemma!"}'



