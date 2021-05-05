#!/bin/bash

set -e

log() {
    echo "`date '+%Y-%m-%d %H:%M:%S'`: $@"
}

if [ -z $ARGO_CLOUDOPS_BUILD_BUCKET ]; then
    echo "ARGO_CLOUDOPS_BUILD_BUCKET not set"
    exit 1
fi

code_dir=app-terraform
build_code_uri="s3://$ARGO_CLOUDOPS_BUILD_BUCKET/terraform-example.tar.gz"
tmp_build=$TMPDIR/terraform-example.tar.gz

log "Creating local archive '$tmp_build'."
rm -f $tmp_build
pushd $code_dir > /dev/null && tar -cvz --exclude='terraform.tfstate*' --exclude='.terraform*' -f $tmp_build .  && popd > /dev/null
log "Creating archive complete."

log "Uploading code to '$build_code_uri'."
aws s3 cp --quiet $tmp_build $build_code_uri
log "Upload code complete."
