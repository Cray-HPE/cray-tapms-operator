/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha3

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GlobalTenantHook) DeepCopyInto(out *GlobalTenantHook) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GlobalTenantHook.
func (in *GlobalTenantHook) DeepCopy() *GlobalTenantHook {
	if in == nil {
		return nil
	}
	out := new(GlobalTenantHook)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GlobalTenantHook) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GlobalTenantHookList) DeepCopyInto(out *GlobalTenantHookList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GlobalTenantHook, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GlobalTenantHookList.
func (in *GlobalTenantHookList) DeepCopy() *GlobalTenantHookList {
	if in == nil {
		return nil
	}
	out := new(GlobalTenantHookList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GlobalTenantHookList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HookCredentials) DeepCopyInto(out *HookCredentials) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HookCredentials.
func (in *HookCredentials) DeepCopy() *HookCredentials {
	if in == nil {
		return nil
	}
	out := new(HookCredentials)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HsmComponent) DeepCopyInto(out *HsmComponent) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HsmComponent.
func (in *HsmComponent) DeepCopy() *HsmComponent {
	if in == nil {
		return nil
	}
	out := new(HsmComponent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HsmComponentList) DeepCopyInto(out *HsmComponentList) {
	*out = *in
	if in.Components != nil {
		in, out := &in.Components, &out.Components
		*out = make([]HsmComponent, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HsmComponentList.
func (in *HsmComponentList) DeepCopy() *HsmComponentList {
	if in == nil {
		return nil
	}
	out := new(HsmComponentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HsmGroup) DeepCopyInto(out *HsmGroup) {
	*out = *in
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.Members.DeepCopyInto(&out.Members)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HsmGroup.
func (in *HsmGroup) DeepCopy() *HsmGroup {
	if in == nil {
		return nil
	}
	out := new(HsmGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HsmIds) DeepCopyInto(out *HsmIds) {
	*out = *in
	if in.Ids != nil {
		in, out := &in.Ids, &out.Ids
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HsmIds.
func (in *HsmIds) DeepCopy() *HsmIds {
	if in == nil {
		return nil
	}
	out := new(HsmIds)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HsmMemberId) DeepCopyInto(out *HsmMemberId) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HsmMemberId.
func (in *HsmMemberId) DeepCopy() *HsmMemberId {
	if in == nil {
		return nil
	}
	out := new(HsmMemberId)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HsmPartition) DeepCopyInto(out *HsmPartition) {
	*out = *in
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.Members.DeepCopyInto(&out.Members)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HsmPartition.
func (in *HsmPartition) DeepCopy() *HsmPartition {
	if in == nil {
		return nil
	}
	out := new(HsmPartition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KeycloakGroup) DeepCopyInto(out *KeycloakGroup) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KeycloakGroup.
func (in *KeycloakGroup) DeepCopy() *KeycloakGroup {
	if in == nil {
		return nil
	}
	out := new(KeycloakGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KeycloakRole) DeepCopyInto(out *KeycloakRole) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KeycloakRole.
func (in *KeycloakRole) DeepCopy() *KeycloakRole {
	if in == nil {
		return nil
	}
	out := new(KeycloakRole)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Tenant) DeepCopyInto(out *Tenant) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tenant.
func (in *Tenant) DeepCopy() *Tenant {
	if in == nil {
		return nil
	}
	out := new(Tenant)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Tenant) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantEventPayload) DeepCopyInto(out *TenantEventPayload) {
	*out = *in
	in.TenantSpec.DeepCopyInto(&out.TenantSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantEventPayload.
func (in *TenantEventPayload) DeepCopy() *TenantEventPayload {
	if in == nil {
		return nil
	}
	out := new(TenantEventPayload)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantHook) DeepCopyInto(out *TenantHook) {
	*out = *in
	if in.EventTypes != nil {
		in, out := &in.EventTypes, &out.EventTypes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.HookCredentials = in.HookCredentials
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantHook.
func (in *TenantHook) DeepCopy() *TenantHook {
	if in == nil {
		return nil
	}
	out := new(TenantHook)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantKmsResource) DeepCopyInto(out *TenantKmsResource) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantKmsResource.
func (in *TenantKmsResource) DeepCopy() *TenantKmsResource {
	if in == nil {
		return nil
	}
	out := new(TenantKmsResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantKmsStatus) DeepCopyInto(out *TenantKmsStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantKmsStatus.
func (in *TenantKmsStatus) DeepCopy() *TenantKmsStatus {
	if in == nil {
		return nil
	}
	out := new(TenantKmsStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantList) DeepCopyInto(out *TenantList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Tenant, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantList.
func (in *TenantList) DeepCopy() *TenantList {
	if in == nil {
		return nil
	}
	out := new(TenantList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TenantList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantResource) DeepCopyInto(out *TenantResource) {
	*out = *in
	if in.Xnames != nil {
		in, out := &in.Xnames, &out.Xnames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantResource.
func (in *TenantResource) DeepCopy() *TenantResource {
	if in == nil {
		return nil
	}
	out := new(TenantResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantSpec) DeepCopyInto(out *TenantSpec) {
	*out = *in
	if in.ChildNamespaces != nil {
		in, out := &in.ChildNamespaces, &out.ChildNamespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.TenantResources != nil {
		in, out := &in.TenantResources, &out.TenantResources
		*out = make([]TenantResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.TenantKmsResource = in.TenantKmsResource
	if in.TenantHooks != nil {
		in, out := &in.TenantHooks, &out.TenantHooks
		*out = make([]TenantHook, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantSpec.
func (in *TenantSpec) DeepCopy() *TenantSpec {
	if in == nil {
		return nil
	}
	out := new(TenantSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantStatus) DeepCopyInto(out *TenantStatus) {
	*out = *in
	if in.ChildNamespaces != nil {
		in, out := &in.ChildNamespaces, &out.ChildNamespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.TenantResources != nil {
		in, out := &in.TenantResources, &out.TenantResources
		*out = make([]TenantResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.TenantKmsStatus = in.TenantKmsStatus
	if in.TenantHooks != nil {
		in, out := &in.TenantHooks, &out.TenantHooks
		*out = make([]TenantHook, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantStatus.
func (in *TenantStatus) DeepCopy() *TenantStatus {
	if in == nil {
		return nil
	}
	out := new(TenantStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Xnames) DeepCopyInto(out *Xnames) {
	{
		in := &in
		*out = make(Xnames, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Xnames.
func (in Xnames) DeepCopy() Xnames {
	if in == nil {
		return nil
	}
	out := new(Xnames)
	in.DeepCopyInto(out)
	return *out
}
