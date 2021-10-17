# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.191.1/containers/ubuntu/.devcontainer/base.Dockerfile

# [Choice] Ubuntu version: bionic, focal
ARG VARIANT="focal"
FROM mcr.microsoft.com/vscode/devcontainers/base:0-${VARIANT}

RUN apt update -y && \
    apt install docker.io -y && \
    sudo apt install bash-completion -y

RUN sudo apt install golang-1.16 -y

RUN sudo apt install make -y

RUN sudo rm -rf /usr/bin/hd && \
    curl -L https://github.com/linuxsuren/http-downloader/releases/latest/download/hd-linux-amd64.tar.gz | tar xzv && \
    mv hd /usr/local/bin && \
    hd fetch && \
    hd install cli/cli && \
    hd install ks && \
    hd install minikube && \
    hd install helm

RUN ks alias init && \
    ks completion bash > /etc/bash_completion.d/ks

RUN echo "export PATH=/usr/lib/go-1.16/bin:$PATH" >> /etc/bash.bashrc && \
    /usr/lib/go-1.16/bin/go get -v golang.org/x/tools/gopls

RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
