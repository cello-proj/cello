DOCKER_HUB_USER ?= celloproj
CDK_REPO := ${DOCKER_HUB_USER}/cello-cdk
TERRAFORM_REPO := ${DOCKER_HUB_USER}/cello-terraform

CDK_VERSION := 1.99.0
TERRAFORM_VERSION := 0.15.1

all: cdk terraform

cdk:
	@echo "Building cdk image."
	cd cdk/ && bash build.sh $(CDK_VERSION) $(CDK_REPO)

terraform:
	@echo "Building terraform image."
	cd terraform/ && bash build.sh $(TERRAFORM_VERSION) $(TERRAFORM_REPO)

.PHONY: cdk terraform
