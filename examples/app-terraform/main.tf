data "aws_region" "current" {}

terraform {
  backend "s3" {
    bucket = "this-is-some-tf-state-for-argo-cloudops"
    key    = "tfstate"
    region = "us-west-2"
  }
}

resource "aws_sqs_queue" "my-example-queue" {
  name = "my-example-queue"
}
