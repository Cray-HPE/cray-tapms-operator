# TAPMS Tenant Status API
Read-Only APIs to Retrieve Tenant Status

## Version: v1alpha3

---
### /v1alpha3/tenants

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

#### POST
##### Summary

Get list of tenants' spec/status with xname ownership

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| xnames | body | Array of Xnames | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [ [Tenant](#tenant) ] |
| 400 | Bad Request | [ResponseError](#responseerror) |
| 404 | Not Found | [ResponseError](#responseerror) |
| 500 | Internal Server Error | [ResponseError](#responseerror) |

### /v1alpha3/tenants/{id}

#### GET
##### Summary

Get a tenant's spec/status

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path | Either the Name or UUID of the Tenant | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [Tenant](#tenant) |
| 400 | Bad Request | [ResponseError](#responseerror) |
| 404 | Not Found | [ResponseError](#responseerror) |
| 500 | Internal Server Error | [ResponseError](#responseerror) |

---
### Models

#### ResponseError

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string | *Example:* `"Error Message..."` | No |

#### Tenant

The primary schema/definition of a tenant

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| spec | [TenantSpec](#tenantspec) | The desired state of Tenant | Yes |
| status | [TenantStatus](#tenantstatus) | The observed state of Tenant | No |

#### TenantHook

The webhook definition to call an API for tenant CRUD operations

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| blockingcall | boolean | +kubebuilder:default:=false +kubebuilder:validation:Optional | No |
| eventtypes | [ string ] | *Example:* `["CREATE"," UPDATE"," DELETE"]` | No |
| name | string |  | No |
| url | string | *Example:* `"http://<url>:<port>"` | No |

#### TenantKmsResource

The Vault KMS transit engine specification for the tenant

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| enablekms | boolean | +kubebuilder:default:=false +kubebuilder:validation:Optional Create a Vault transit engine for the tenant if this setting is true. | No |
| keyname | string | +kubebuilder:default:=key1 +kubebuilder:validation:Optional Optional name for the transit engine key. | No |
| keytype | string | +kubebuilder:default:=rsa-3072 +kubebuilder:validation:Optional Optional key type. See https://developer.hashicorp.com/vault/api-docs/secret/transit#type The default of 3072 is the minimal permitted under the Commercial National Security Algorithm (CNSA) 1.0 suite. | No |

#### TenantKmsStatus

The Vault KMS transit engine status for the tenant

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| keyname | string | The Vault transit key name. | No |
| keytype | string | The Vault transit key type. | No |
| publickey | string | The Vault public key. | No |
| transitname | string | The generated Vault transit engine name. | No |

#### TenantResource

The desired resources for the Tenant

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| enforceexclusivehsmgroups | boolean |  | No |
| hsmgrouplabel | string | *Example:* `"green"` | No |
| hsmpartitionname | string | *Example:* `"blue"` | No |
| type | string | *Example:* `"compute"` | Yes |
| xnames | [ string ] | *Example:* `["x0c3s5b0n0","x0c3s6b0n0"]` | Yes |

#### TenantSpec

The desired state of Tenant

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| childnamespaces | [ string ] | *Example:* `["vcluster-blue-slurm"]` | No |
| state | string | +kubebuilder:validation:Optional<br>*Example:* `"New,Deploying,Deployed,Deleting"` | No |
| tenanthooks | [ [TenantHook](#tenanthook) ] | +kubebuilder:validation:Optional | No |
| tenantkms | [TenantKmsResource](#tenantkmsresource) | +kubebuilder:validation:Optional | No |
| tenantname | string | *Example:* `"vcluster-blue"` | Yes |
| tenantresources | [ [TenantResource](#tenantresource) ] | The desired resources for the Tenant | Yes |

#### TenantStatus

The observed state of Tenant

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| childnamespaces | [ string ] | *Example:* `["vcluster-blue-slurm"]` | No |
| tenanthooks | [ [TenantHook](#tenanthook) ] |  | No |
| tenantkms | [TenantKmsStatus](#tenantkmsstatus) |  | No |
| tenantresources | [ [TenantResource](#tenantresource) ] | The desired resources for the Tenant | No |
| uuid | string (uuid) | *Example:* `"550e8400-e29b-41d4-a716-446655440000"` | No |
