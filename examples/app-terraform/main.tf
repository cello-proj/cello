data "aws_region" "current" {}

terraform {
  backend "local" {
    path = "terraform.tfstate"
  }
}

resource "aws_sqs_queue" "my-example-queue" {
  name = "my-example-queue"
}
