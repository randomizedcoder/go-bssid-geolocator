#!/bin/bash

# sudo because the output is written as root from docker
sudo rm -rf ./pkg/bssid

docker run --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf dep update
docker run --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf lint
docker run --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf build
docker run --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf generate

# USER_ID=$(id -u)
# GROUP_ID=$(id -g)

# sudo chown -R "${USER_ID}:${GROUP_ID}" ./pkg/bssid

# end
