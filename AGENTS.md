## container-platform-tools

It provides two commands implemented in Go:

- `docker-platforms` lists supported platforms for the given OCI image.
- `kubectl-platforms` lists supported platforms for workloads in the k8s cluster.

## Structure

- `./cmd` - contains the two commands
- `./dockerplatforms` - the core of ./cmd/docker-platforms
- `./k8splatforms` - the core of ./cmd/kubectl-platforms
