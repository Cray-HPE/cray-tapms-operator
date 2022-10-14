
### /apis/tapms/v1/tenant/{id}

#### GET
##### Summary

Get a tenant spec/status

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path | either name or uuid of the tenant | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [ResponseOk](#responseok) |
| 400 | Bad Request | [ResponseError](#responseerror) |
| 404 | Not Found | [ResponseError](#responseerror) |
| 500 | Internal Server Error | [ResponseError](#responseerror) |

### /apis/tapms/v1/tenants

#### GET
##### Summary

Get list of tenants

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [ResponseOk](#responseok) |
| 400 | Bad Request | [ResponseError](#responseerror) |
| 404 | Not Found | [ResponseError](#responseerror) |
| 500 | Internal Server Error | [ResponseError](#responseerror) |

### Models

#### ResponseError

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |

#### ResponseOk

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |
