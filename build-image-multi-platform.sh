#!/bin/bash
set -e

NAME=$1
DOCKERFILE=$2
TAG=$3
PLATFORM=${4:-linux/amd64,linux/arm64}  # Default platform is linux/amd64 and linux/arm64 if not specified

if [ $TAG ]; then
   echo "Building dockerfile $DOCKERFILE for $NAME:$TAG on platforms $PLATFORM"
   docker buildx build --rm=false --file ./$DOCKERFILE --platform $PLATFORM -t $NAME:$TAG ./. --push
else
   echo "Building dockerfile $DOCKERFILE for $NAME on platforms $PLATFORM"
   docker buildx build --rm=false --file ./$DOCKERFILE --platform $PLATFORM -t $NAME ./. --push
   docker tag $NAME $NAME:latest
   docker tag $NAME $NAME:$BUILDPREFIX$GITHUB_RUN_NUMBER
fi