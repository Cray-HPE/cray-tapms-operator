/*
 *
 *  MIT License
 *
 *  (C) Copyright 2022 Hewlett Packard Enterprise Development LP
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

package v1alpha1

import (
	"errors"

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

//+kubebuilder:webhook:path=/mutate-tapms-hpe-com-v1alpha1-tenant,mutating=true,failurePolicy=fail,sideEffects=None,groups=tapms.hpe.com,resources=tenants,verbs=create;update,versions=v1alpha1,name=mtenant.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Tenant{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (t *Tenant) Default() {
	Log.Info("Validating default for", "tenant", t.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-tapms-hpe-com-v1alpha1-tenant,mutating=false,failurePolicy=fail,sideEffects=None,groups=tapms.hpe.com,resources=tenants,verbs=create;update,versions=v1alpha1,name=vtenant.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Tenant{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (t *Tenant) ValidateCreate() error {
	Log.Info("Validating create for", "tenant", t.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (t *Tenant) ValidateUpdate(old runtime.Object) error {
	Log.Info("Validating update for", "tenant", t.Name)
	for _, specResource := range t.Spec.TenantResources {
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

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (t *Tenant) ValidateDelete() error {
	Log.Info("Validating delete for", "tenant", t.Name)
	return nil
}
