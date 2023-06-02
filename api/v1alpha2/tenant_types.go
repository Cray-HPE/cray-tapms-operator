/*
 *
 *  MIT License
 *
 *  (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP
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
/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// @Description The desired resources for the Tenant
type TenantResource struct {
	Type                      string   `json:"type" example:"compute" binding:"required"`
	Xnames                    []string `json:"xnames" example:"x0c3s5b0n0,x0c3s6b0n0" binding:"required"`
	HsmPartitionName          string   `json:"hsmpartitionname,omitempty" example:"blue"`
	HsmGroupLabel             string   `json:"hsmgrouplabel,omitempty" example:"green"`
	EnforceExclusiveHsmGroups bool     `json:"enforceexclusivehsmgroups"`
	//+kubebuilder:validation:Optional
	ForcePowerOff bool `json:"forcepoweroff"`
} // @name TenantResource

// The Vault KMS transit engine specification for the tenant
type TenantKmsResource struct {
	//+kubebuilder:default:=false
	//+kubebuilder:validation:Optional
	// Create a Vault transit engine for the tenant if this setting is true.
	Enabled bool `json:"enablekms"`
	//+kubebuilder:default:=key1
	//+kubebuilder:validation:Optional
	// Optional name for the transit engine key.
	KeyName string `json:"keyname"`
	//+kubebuilder:default:=rsa-2048
	//+kubebuilder:validation:Optional
	// Optional key type.
	//See https://developer.hashicorp.com/vault/api-docs/secret/transit#type
	KeyType string `json:"keytype"`
}

// The Vault KMS transit engine status for the tenant
type TenantKmsStatus struct {
	// The generated Vault transit engine name.
	TransitName string `json:"transitname"`
	// The Vault transit key name.
	KeyName string `json:"keyname"`
	// The Vault transit key type.
	KeyType string `json:"keytype"`
}

// @Description The desired state of Tenant
type TenantSpec struct {
	TenantName string `json:"tenantname" example:"vcluster-blue" binding:"required"`
	//+kubebuilder:validation:Optional
	State           string   `json:"state" example:"New,Deploying,Deployed,Deleting"`
	ChildNamespaces []string `json:"childnamespaces" example:"vcluster-blue-slurm"`
	// The desired resources for the Tenant
	TenantResources []TenantResource `json:"tenantresources" binding:"required"`
	//+kubebuilder:validation:Optional
	TenantKmsResource TenantKmsResource `json:"tenantkms"`
} //@name TenantSpec

// @Description The observed state of Tenant
type TenantStatus struct {
	ChildNamespaces []string `json:"childnamespaces,omitempty" example:"vcluster-blue-slurm"`
	// The desired resources for the Tenant
	TenantResources []TenantResource `json:"tenantresources,omitempty"`
	UUID            string           `json:"uuid,omitempty" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	TenantKmsStatus TenantKmsStatus  `json:"tenantkms,omitempty"`
} // @name TenantStatus

//+k8s:openapi-gen=true
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// @Description The primary schema/definition of a tenant
type Tenant struct {
	metav1.TypeMeta   `json:",inline" swaggerignore:"true"`
	metav1.ObjectMeta `json:"metadata,omitempty" swaggerignore:"true"`
	// The desired state of Tenant
	Spec TenantSpec `json:"spec,omitempty" binding:"required"`
	// The observed state of Tenant
	Status TenantStatus `json:"status,omitempty"`
} // @name Tenant

//+kubebuilder:object:root=true

// @Description List of Tenants
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}
