#!/bin/bash

set -e

cdk_version=$1
repo=$2

usage() {
    echo "$0 CDK_VERSION REPO"
}

if [ -z $cdk_version ]; then
    usage
    exit 1
fi

if [ -z $repo ]; then
    usage
    exit 1
fi

build_dir=$TMPDIR/docker-cdk

rm -rf $build_dir

mkdir -p $build_dir

cp Dockerfile $build_dir
cp requirements.txt $build_dir
cp ../shared/setup.sh $build_dir

cd $build_dir

sed -i '' "s/{{CDK_VERSION}}/$cdk_version/g" Dockerfile requirements.txt

tags="-t $repo:$cdk_version -t $repo:latest"

docker build . --no-cache $tags

docker push $repo:$cdk_version
docker push $repo:latest
