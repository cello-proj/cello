#!/bin/bash

set -e

terraform_version=$1
repo=$2

usage() {
    echo "$0 TERRAFORM_VERSION REPO"
}

if [ -z $terraform_version ]; then
    usage
    exit 1
fi

if [ -z $repo ]; then
    usage
    exit 1
fi

build_dir=$TMPDIR/docker-terraform

\rm -rf $build_dir

mkdir -p $build_dir

cp Dockerfile $build_dir
cp requirements.txt $build_dir
cp ../shared/setup.sh $build_dir

cd $build_dir

sed -i '' "s/{{TERRAFORM_VERSION}}/$terraform_version/g" Dockerfile

tags="-t $repo:$terraform_version -t $repo:latest"

docker build . --no-cache $tags

docker push $repo:$terraform_version
docker push $repo:latest
