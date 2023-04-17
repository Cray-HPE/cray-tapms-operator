# TAPMS Tenant Status API
Read-Only APIs to Retrieve Tenant Status

## Version: v1alpha1

### /v1/tenants

#### GET
##### Summary

Get list of tenants' spec/status

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [ [Tenant](#tenant) ] |
| 400 | Bad Request | [ResponseError](#responseerror) |
| 404 | Not Found | [ResponseError](#responseerror) |
| 500 | Internal Server Error | [ResponseError](#responseerror) |

### /v1/tenants/{id}

#### GET
##### Summary

Get a tenant's spec/status

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path | Either the Name or UUID of the Tenant | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [Tenant](#tenant) |
| 400 | Bad Request | [ResponseError](#responseerror) |
| 404 | Not Found | [ResponseError](#responseerror) |
| 500 | Internal Server Error | [ResponseError](#responseerror) |

### Models

#### ResponseError

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string | _Example:_ `"Error Message..."` | No |

#### Tenant

The primary schema/definition of a tenant

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| spec | [TenantSpec](#tenantspec) | The desired state of Tenant | Yes |
| status | [TenantStatus](#tenantstatus) | The observed state of Tenant | No |

#### TenantResource

The desired resources for the Tenant

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| enforceexclusivehsmgroups | boolean |  | No |
| hsmgrouplabel | string | _Example:_ `"green"` | No |
| hsmpartitionname | string | _Example:_ `"blue"` | No |
| type | string | _Example:_ `"compute"` | Yes |
| xnames | [ string ] | _Example:_ `["x0c3s5b0n0","x0c3s6b0n0"]` | Yes |

#### TenantSpec

The desired state of Tenant

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| childnamespaces | [ string ] | _Example:_ `["vcluster-blue-slurm"]` | No |
| state | string | _Example:_ `"New,Deploying,Deployed,Deleting"` | No |
| tenantname | string | _Example:_ `"vcluster-blue"` | Yes |
| tenantresources | [ [TenantResource](#tenantresource) ] | The desired resources for the Tenant | Yes |

#### TenantStatus

The observed state of Tenant

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| childnamespaces | [ string ] | _Example:_ `["vcluster-blue-slurm"]` | No |
| tenantresources | [ [TenantResource](#tenantresource) ] | The desired resources for the Tenant | No |
| uuid | string (uuid) | _Example:_ `"550e8400-e29b-41d4-a716-446655440000"` | No |
