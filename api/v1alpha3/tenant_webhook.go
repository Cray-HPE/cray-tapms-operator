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

package v1alpha3

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var Log = logf.Log.WithName("tenants")

func (t *Tenant) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(t).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-tapms-hpe-com-v1alpha3-tenant,mutating=true,failurePolicy=fail,sideEffects=None,groups=tapms.hpe.com,resources=tenants,verbs=create;update,versions=v1alpha3,name=mtenant.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Tenant{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (t *Tenant) Default() {
	Log.Info("Validating default for", "tenant", t.Name)

	if t.GetDeletionTimestamp() != nil {
		t.Spec.State = "Deleting"
	} else if t.Spec.State == "" {
		t.Spec.State = "New"
	} else if TenantIsUpdated(t) {
		t.Spec.State = "Deploying"
	} else {
		t.Spec.State = "Deployed"
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-tapms-hpe-com-v1alpha3-tenant,mutating=false,failurePolicy=fail,sideEffects=None,groups=tapms.hpe.com,resources=tenants,verbs=create;update;delete,versions=v1alpha3,name=vtenant.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Tenant{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (t *Tenant) ValidateCreate() error {
	Log.Info("Validating create for", "tenant", t.Name)

	for _, specResource := range t.Spec.TenantResources {
		if specResource.Type == "compute" {
			err := t.ValidateNodeTypeForXnames(specResource.Xnames, "Node", "Compute")
			if err != nil {
				return err
			}

		}
		if specResource.Type == "application" {
			err := t.ValidateNodeTypeForXnames(specResource.Xnames, "Node", "Application")
			if err != nil {
				return err
			}

		}
		err := t.ValidateExclusiveGroupMembership(specResource.Xnames, specResource.HsmGroupLabel, specResource.EnforceExclusiveHsmGroups)
		if err != nil {
			return err
		}
	}

	err := CallHooks(t, Log, "CREATE")
	if err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (t *Tenant) ValidateUpdate(old runtime.Object) error {
	Log.Info("Validating update for", "tenant", t.Name)

	for _, specResource := range t.Spec.TenantResources {

		if specResource.Type == "compute" {
			err := t.ValidateNodeTypeForXnames(specResource.Xnames, "Node", "Compute")
			if err != nil {
				return err
			}

		}
		if specResource.Type == "application" {
			err := t.ValidateNodeTypeForXnames(specResource.Xnames, "Node", "Application")
			if err != nil {
				return err
			}
		}

		err := t.ValidateExclusiveGroupMembership(specResource.Xnames, specResource.HsmGroupLabel, specResource.EnforceExclusiveHsmGroups)
		if err != nil {
			return err
		}

		for _, statusResource := range t.Status.TenantResources {
			if statusResource.Type == specResource.Type {
				//
				// Various update validations here
				//
				if statusResource.EnforceExclusiveHsmGroups != specResource.EnforceExclusiveHsmGroups {
					//
					// EnforceExclusiveHsmGroups is immutable as we have no API support
					// from HSM for changing the setting after HSM group is created.
					//
					err := errors.New("EnforceExclusiveHsmGroups field is immutable")
					Log.Error(err, "Failed to update tenant")
					return err
				}
			}
		}
	}

	err := CallHooks(t, Log, "UPDATE")
	if err != nil {
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (t *Tenant) ValidateDelete() error {
	Log.Info("Validating delete for", "tenant", t.Name)
	err := CallHooks(t, Log, "DELETE")
	if err != nil {
		return err
	}
	return nil
}

func (t *Tenant) ValidateNodeTypeForXnames(xnames []string, nodeType string, role string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hsmComponentList, err := GetComponentList(ctx, Log, nodeType, role)
	if err != nil {
		return err
	}
	var failedXnames []string
	failedXnames = make([]string, 0, len(xnames))

	for _, xname := range xnames {
		found := false
		for _, component := range hsmComponentList.Components {
			if xname == component.ID {
				found = true
				break
			}
		}
		if !found {
			failedXnames = append(failedXnames, xname)
		}
	}

	if len(failedXnames) > 0 {
		return fmt.Errorf("the following xname(s) do not have type %s and role %s: %v", nodeType, role, failedXnames)
	}

	return nil
}

func (t *Tenant) ValidateExclusiveGroupMembership(xnames []string, hsmGroupLabel string, exclusiveFlag bool) error {

	if !exclusiveFlag {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, groupList, err := ListHSMGroups(ctx, Log)
	if err != nil {
		return err
	}
	var currentTenantXnames []string
	currentTenantXnames = make([]string, 0, len(xnames))
	for _, xname := range xnames {
		for _, group := range groupList {
			if (group.ExclusiveGroup != "tapms-exclusive-group-label") || (group.Label == hsmGroupLabel) {
				continue
			}

			for _, member := range group.Members.Ids {
				if xname == member {
					currentTenantXnames = append(currentTenantXnames, xname)
				}
			}
		}
	}

	if len(currentTenantXnames) > 0 {
		return fmt.Errorf("the following xname(s): '%v' already exist in an exclusive hsm group", currentTenantXnames)
	}

	return nil
}
