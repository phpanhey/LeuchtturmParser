#!/bin/bash

#cross compile for linux
GOOS=linux GOARCH=amd64 go build -o ./bin/release/linux/LeuchtturmParser

# Load environment variables from .env file
source .env

# Copy the LeuchtturmParser file to the destination server using scp
scp bin/release/linux/* $REMOTE_SERVER_USER@$REMOTE_SERVER_IP:$REMOTE_SERVER_DIR
