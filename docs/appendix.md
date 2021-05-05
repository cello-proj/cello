# Appendix

## Artifact Passing Between Steps

Only necessary if passing artifacts between steps, not required by default.

* Install [Minio](https://argoproj.github.io/argo-workflows/configure-artifact-repository/#configuring-minio) for internal K8s artifact storage.
* Setup [Minio](https://gist.github.com/bw-intuit/62bd62b5d4eb9088572ae261d8dfea1a#file-miniosetupforargodockerdesktop-md) for Argo.
* Ensure sample [artifact passing workflow](https://raw.githubusercontent.com/argoproj/argo/master/examples/artifact-passing.yaml) works.
