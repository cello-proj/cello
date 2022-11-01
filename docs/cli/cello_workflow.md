## cello workflow
Creates a workflow execution with provided arguments

```
  cello workflow [flags]
```

### Flags

```
  -a, --arguments string                CSV string of equals separated arguments to pass to command (-a Arg1=ValueA,Arg2=ValueB).
  -e, --environment_variables string    CSV string of equals separated environment variable key value pairs (-e Key1=ValueA,Key2=ValueB)
  -f, --framework string                Framework to execute
  -h, --help                            help for workflow
  -p, --parameters string               CSV string of equals separated parameters name and value (-p Param1=ValueA,Param2=ValueB).
  -n, --project_name string             Name of project
  -t, --target string                   Name of target
      --type string                     Workflow type to execute
  -w, --workflow_template_name string   Name of the workflow template
```