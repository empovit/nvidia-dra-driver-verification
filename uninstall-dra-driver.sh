#!/bin/sh

set -ex

helm uninstall --namespace nvidia-dra-driver-gpu nvidia-dra-driver-gpu
