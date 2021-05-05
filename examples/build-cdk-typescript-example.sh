#!/bin/bash

set -e

log() {
    echo "`date '+%Y-%m-%d %H:%M:%S'`: $@"
}

if [ -z $ARGO_CLOUDOPS_BUILD_BUCKET ]; then
    echo "ARGO_CLOUDOPS_BUILD_BUCKET not set"
    exit 1
fi

code_dir=app-cdk-typescript
build_code_uri="s3://$ARGO_CLOUDOPS_BUILD_BUCKET/cdk-typescript-example.tar.gz"
tmp_build=$TMPDIR/cdk-typescript-example.tar.gz

log "Creating CDK code archive."
rm -f $tmp_build
pushd $code_dir > /dev/null

log "Running 'npm install'."
npm install

log "Running 'npm install' complete."
tar -czf $tmp_build .
popd > /dev/null
log "Creating archive complete."

log "Uploading code to '$build_code_uri'."
aws s3 cp --quiet $tmp_build $build_code_uri
log "Upload code complete."
