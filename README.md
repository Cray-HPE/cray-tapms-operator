# cray-tapms-operator (Tenant and Partition Management System)

## Overview

This chart is responsible for multi-tenancy operations for CSM.  Installing this chart deploys a Kubernetes CRD (Custom Resource Definition) through which tenant operations are managed.  See section below for example tenant creations.

Deploying this chart will create the cray-tapms-operator Kubernetes operator/controller following the Operator Pattern (https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

This chart leverages the Hierarchical Namespace Controller (HNC) for namespace management, and can apply policies like RBAC and Network Policies across all namespaces for a given tenant.  For more information about HNC, see https://github.com/kubernetes-sigs/hierarchical-namespaces.

## Project Status: alpha

Initial implementation of this chart should be considered `soft` multi-tenancy, with improvements coming in future versions.

## Sample CRD (Custom Resource Definition)

See [example yaml](./config/samples/tapms.hpe.com_v1alpha1_tenant.yaml) for an example tenant specification.

## Create a Tenant

The `kubectl` command (and K8S API) can be used to create a tenant:

```
% kubectl -n tenants -f tenant.yaml apply
  tenant.tapms.hpe.com/tenant-dev created
```

## Update a Tenant

Similarly, `kubectl` command (and K8S API) can be used to update a tenant:

```
% kubectl -n tenants -f tenant.yaml apply
  tenant.tapms.hpe.com/tenant-dev changed
```

## Destroy a Tenant

Finally, `kubectl` command (and K8S API) can be used to delete/remove a tenant:

```
% kubectl -n tenants -f tenant.yaml delete
  tenant.tapms.hpe.com/tenant-dev deleted
```

## View HNC Tenant Structure

CSM NCNs are deployed with the `kubectl-hns` plugin, and as such when logged into an NCN (master, worker, storage node), the following command may be useful to display a tree view of tenants:

```
% kubectl hns tree tenants
  tenants
  └── [s] tenant-dev
      ├── [s] tenant-dev-slurm
      └── [s] tenant-dev-user
```

## Update swagger

   ```
   scripts/swagger.gen.sh
   ```
   > Note: This script will try to update `docs/swagger.md` if nodejs is installed. Otherwise, it will only update `docs/swagger.yaml`.  [NodeJS](https://nodejs.org/en/download/) is required for markdown version of swagger doc.

## Changing CRD/API version notes

Below are the developer steps for altering the Tenant CRD and API version, see (Kubernetes Changing the API)[https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md]:

1. First copy apis/&lt;existing-version&gt; to apis/&lt;new-version&gt;.
1. Rename &lt;existing-version&gt; to &lt;new-version&gt; in the new files.
1. Make code changes necessary for &lt;new-version&gt;.
1. Bump appropriate chart/docker versions.
1. Generate files by running make manifests/schema (CRD gets generated).
1. Run 'make charts' to ensure the charts build and the new CRD moves to kubernetes/cray-tapms-crd/files.
1. Run 'scripts/swagger.gen.sh' to update the swagger/openapi spec.
1. Convert the swagger.yaml to the openapi.yaml spec using https://editor.swagger.io/ until we improve that process.
