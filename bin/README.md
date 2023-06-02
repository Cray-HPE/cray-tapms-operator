# Setup
Operator development and the `Makefile` for this project can require certain tools.

## Installing the `operator-sdk` tool
Ensure that you have installed the Operator SDK tool. Follow the process [here](https://sdk.operatorframework.io/docs/installation/). If on OSX, it can be installed using `brew install operator-sdk`, but review the installation page for options. The `brew` method can also fail if your OSX version is too old. If you choose to install from a repo, note that the gpg instructions may not work.  

## Getting and installing `bin/controller-gen`
Depending on the task, you may need to install `controller-gen`. Note that this is created from the Makefile when a new operator project is generated but not checked in.
Review the `go.mod` and look for the controller runtime. Example: `sigs.k8s.io/controller-runtime v0.11.0`
If the controller version is missing from the Go cache, run `go mod tidy`. Use `go env` to find the location of your cache in `GOMODCACHE`. 
In this example, from the main project directory, build `controller-gen` by running:

```bash
GOBIN=$PWD/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.11.1
```
