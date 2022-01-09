examples: cdk-typescript-example terraform-example

cdk-typescript-example:
	@echo "Building cdk-typescript-example."
	env ARGO_CLOUDOPS_BUILD_BUCKET=${ARGO_CLOUDOPS_BUILD_BUCKET} bash build-cdk-typescript-example.sh

terraform-example:
	@echo "Building terraform-example."
	env ARGO_CLOUDOPS_BUILD_BUCKET=${ARGO_CLOUDOPS_BUILD_BUCKET} bash build-terraform-example.sh

.PHONY: cdk-typescript-example terraform-example
