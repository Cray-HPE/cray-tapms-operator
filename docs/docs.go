/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/v1alpha3/tenants": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tenant and Partition Management System"
                ],
                "summary": "Get list of tenants' spec/status",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/Tenant"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ResponseError"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/ResponseError"
                        }
                    }
                }
            },
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tenant and Partition Management System"
                ],
                "summary": "Get list of tenants' spec/status with xname ownership",
                "parameters": [
                    {
                        "description": "Array of Xnames",
                        "name": "xnames",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "string",
                            "example": "[\"x1000c0s0b0n0\", \"x1000c0s0b1n0\"]"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/Tenant"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ResponseError"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/ResponseError"
                        }
                    }
                }
            }
        },
        "/v1alpha3/tenants/{id}": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tenant and Partition Management System"
                ],
                "summary": "Get a tenant's spec/status",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Either the Name or UUID of the Tenant",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/Tenant"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ResponseError"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/ResponseError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "HookCredentials": {
            "description": "Optional credentials for calling webhook",
            "type": "object",
            "properties": {
                "secretname": {
                    "description": "+kubebuilder:validation:Optional\nOptional Kubernetes secret name containing credentials for calling webhook",
                    "type": "string"
                },
                "secretnamespace": {
                    "description": "+kubebuilder:validation:Optional\nOptional Kubernetes namespace for the secret",
                    "type": "string"
                }
            }
        },
        "ResponseError": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string",
                    "example": "Error Message..."
                }
            }
        },
        "Tenant": {
            "description": "The primary schema/definition of a tenant",
            "type": "object",
            "required": [
                "spec"
            ],
            "properties": {
                "spec": {
                    "description": "The desired state of Tenant",
                    "allOf": [
                        {
                            "$ref": "#/definitions/TenantSpec"
                        }
                    ]
                },
                "status": {
                    "description": "The observed state of Tenant",
                    "allOf": [
                        {
                            "$ref": "#/definitions/TenantStatus"
                        }
                    ]
                }
            }
        },
        "TenantHook": {
            "description": "The webhook definition to call an API for tenant CRUD operations",
            "type": "object",
            "properties": {
                "blockingcall": {
                    "description": "+kubebuilder:default:=false\n+kubebuilder:validation:Optional",
                    "type": "boolean"
                },
                "eventtypes": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "CREATE",
                        " UPDATE",
                        " DELETE"
                    ]
                },
                "hookcredentials": {
                    "description": "+kubebuilder:validation:Optional",
                    "allOf": [
                        {
                            "$ref": "#/definitions/HookCredentials"
                        }
                    ]
                },
                "name": {
                    "type": "string"
                },
                "url": {
                    "type": "string",
                    "example": "http://\u003curl\u003e:\u003cport\u003e"
                }
            }
        },
        "TenantKmsResource": {
            "description": "The Vault KMS transit engine specification for the tenant",
            "type": "object",
            "properties": {
                "enablekms": {
                    "description": "+kubebuilder:default:=false\n+kubebuilder:validation:Optional\nCreate a Vault transit engine for the tenant if this setting is true.",
                    "type": "boolean"
                },
                "keyname": {
                    "description": "+kubebuilder:default:=key1\n+kubebuilder:validation:Optional\nOptional name for the transit engine key.",
                    "type": "string"
                },
                "keytype": {
                    "description": "+kubebuilder:default:=rsa-3072\n+kubebuilder:validation:Optional\nOptional key type. See https://developer.hashicorp.com/vault/api-docs/secret/transit#type\nThe default of 3072 is the minimal permitted under the Commercial National Security Algorithm (CNSA) 1.0 suite.",
                    "type": "string"
                }
            }
        },
        "TenantKmsStatus": {
            "description": "The Vault KMS transit engine status for the tenant",
            "type": "object",
            "properties": {
                "keyname": {
                    "description": "The Vault transit key name.",
                    "type": "string"
                },
                "keytype": {
                    "description": "The Vault transit key type.",
                    "type": "string"
                },
                "publickey": {
                    "description": "The Vault public key.",
                    "type": "string"
                },
                "transitname": {
                    "description": "The generated Vault transit engine name.",
                    "type": "string"
                }
            }
        },
        "TenantResource": {
            "description": "The desired resources for the Tenant",
            "type": "object",
            "required": [
                "type",
                "xnames"
            ],
            "properties": {
                "enforceexclusivehsmgroups": {
                    "type": "boolean"
                },
                "hsmgrouplabel": {
                    "type": "string",
                    "example": "green"
                },
                "hsmpartitionname": {
                    "type": "string",
                    "example": "blue"
                },
                "type": {
                    "type": "string",
                    "example": "compute"
                },
                "xnames": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "x0c3s5b0n0",
                        "x0c3s6b0n0"
                    ]
                }
            }
        },
        "TenantSpec": {
            "description": "The desired state of Tenant",
            "type": "object",
            "required": [
                "tenantname",
                "tenantresources"
            ],
            "properties": {
                "childnamespaces": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "vcluster-blue-slurm"
                    ]
                },
                "state": {
                    "description": "+kubebuilder:validation:Optional",
                    "type": "string",
                    "example": "New,Deploying,Deployed,Deleting"
                },
                "tenanthooks": {
                    "description": "+kubebuilder:validation:Optional",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/TenantHook"
                    }
                },
                "tenantkms": {
                    "description": "+kubebuilder:validation:Optional",
                    "allOf": [
                        {
                            "$ref": "#/definitions/TenantKmsResource"
                        }
                    ]
                },
                "tenantname": {
                    "type": "string",
                    "example": "vcluster-blue"
                },
                "tenantresources": {
                    "description": "The desired resources for the Tenant",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/TenantResource"
                    }
                }
            }
        },
        "TenantStatus": {
            "description": "The observed state of Tenant",
            "type": "object",
            "properties": {
                "childnamespaces": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "vcluster-blue-slurm"
                    ]
                },
                "tenanthooks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/TenantHook"
                    }
                },
                "tenantkms": {
                    "$ref": "#/definitions/TenantKmsStatus"
                },
                "tenantresources": {
                    "description": "The desired resources for the Tenant",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/TenantResource"
                    }
                },
                "uuid": {
                    "type": "string",
                    "format": "uuid",
                    "example": "550e8400-e29b-41d4-a716-446655440000"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "v1alpha3",
	Host:             "cray-tapms",
	BasePath:         "/apis/tapms/",
	Schemes:          []string{},
	Title:            "TAPMS Tenant Status API",
	Description:      "Read-Only APIs to Retrieve Tenant Status",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
