#!/bin/bash

# start minikube
sudo chown vscode:vscode /var/run/docker.sock
minikube-linux-amd64 start --ports=30880

# ks init
ks alias init
