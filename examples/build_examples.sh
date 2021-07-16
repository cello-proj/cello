#!/bin/bash

set -e

log() {
    echo "`date '+%Y-%m-%d %H:%M:%S'`: $@"
}

code_dir=app-cdk-typescript
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

code_dir=app-terraform
tmp_build=$TMPDIR/terraform-example.tar.gz

log "Creating local archive '$tmp_build'."
rm -f $tmp_build
pushd $code_dir > /dev/null && tar -cvz --exclude='terraform.tfstate*' --exclude='.terraform*' -f $tmp_build .  && popd > /dev/null
log "Creating archive complete."