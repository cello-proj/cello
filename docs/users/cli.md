# CLI

The CLI allows to (amongst other things) manage projects, sync, watch, and list deployments, e.g.: 

```sh
WFNAME=`argo-cloudops sync -n project1 -t target1 -p git_path -s git_sha`
argo-cloudops logs $WFNAME -f    
```   

## Reference

You can find [detailed reference here](/cli/argo-cloudops)

## Help

Most help topics are provided by built-in help:

```
argo-cloudops --help
```