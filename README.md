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
