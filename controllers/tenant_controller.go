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

package controllers

import (
	"context"
	"fmt"
	"reflect"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/Cray-HPE/cray-tapms-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const tenantFinalizer = "tapms.hpe.com/finalizer"

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tapms.hpe.com,resources=tenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tapms.hpe.com,resources=tenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tapms.hpe.com,resources=tenants/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tenant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("tenants", req.NamespacedName)
	tenant := &v1.Tenant{}
	err := r.Get(ctx, req.NamespacedName, tenant)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Tenant resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
	}

	isTenantMarkedToBeDeleted := tenant.GetDeletionTimestamp() != nil
	if !isTenantMarkedToBeDeleted {
		if v1.TenantIsUpdated(tenant) {
			tenant.Status.UUID = string(tenant.ObjectMeta.UID)
			err = r.Status().Update(ctx, tenant)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to update resource state: %w", err)
			} else {
				err = r.Update(ctx, tenant)
				if err != nil {
					log.Error(err, "Failed to update resource")
					return ctrl.Result{}, err
				}
			}
		}

		result, err := v1.CreateSubanchorNs(ctx, log, r.Client, "tenants", tenant.Spec.TenantName)
		if err != nil {
			return result, err
		} else if result.Requeue {
			return result, nil
		}

		if tenant.Spec.ChildNamespaces != nil {
			for _, childNamespace := range tenant.Spec.ChildNamespaces {
				childNs := v1.GetChildNamespaceName(tenant.Spec.TenantName, childNamespace)
				result, err := v1.CreateSubanchorNs(ctx, log, r.Client, tenant.Spec.TenantName, childNs)
				if err != nil {
					return result, err
				} else if result.Requeue {
					return result, nil
				}
			}
		}

		for _, resource := range tenant.Spec.TenantResources {
			if len(resource.HsmPartitionName) > 0 {
				log.Info(fmt.Sprintf("Creating/updating HSM partition for %s and resource type %s", tenant.Spec.TenantName, resource.Type))
				result, err := v1.UpdateHSMPartition(ctx, log, tenant, resource.HsmPartitionName, resource.Xnames)
				if err != nil {
					log.Error(err, "Failed to create/update HSM partition")
					return result, err
				}
			}
		}

		for _, resource := range tenant.Spec.TenantResources {
			if len(resource.HsmGroupLabel) > 0 {
				log.Info(fmt.Sprintf("Creating/updating HSM group for %s and resource type %s", tenant.Spec.TenantName, resource.Type))
				result, err := v1.UpdateHSMGroup(ctx, log, tenant, resource.HsmGroupLabel, resource.Xnames, resource.EnforceExclusiveHsmGroups)
				if err != nil {
					log.Error(err, "Failed to create/update HSM group")
					return result, err
				}
			}
		}

		log.Info("Creating/updating Keycloak Group for: " + tenant.Spec.TenantName)
		result, err = v1.UpdateKeycloakGroup(ctx, log, tenant)
		if err != nil {
			log.Error(err, "Failed to create/update Keycloak Group")
			return result, err
		}

		if !reflect.DeepEqual(v1.TranslateStatusNamespacesForSpec(tenant.Status.ChildNamespaces), tenant.Spec.ChildNamespaces) {
			//
			// Don't need to add members, that gets handled above in the create loop
			//
			deletedChildNamespaces := v1.Difference(v1.TranslateStatusNamespacesForSpec(tenant.Status.ChildNamespaces), tenant.Spec.ChildNamespaces)
			v1.DeleteChildNamespaces(ctx, log, r.Client, tenant, deletedChildNamespaces)
			if err != nil {
				log.Error(err, "Failed to delete child namespaces")
				return ctrl.Result{}, err
			}
		}

		if v1.TenantIsUpdated(tenant) {
			log.Info("Updating tenant status")
			tenant.Status.TenantResources = tenant.Spec.TenantResources
			tenant.Status.ChildNamespaces = v1.TranslateSpecNamespacesForStatus(tenant.Spec.TenantName, tenant.Spec.ChildNamespaces)
			err = r.Status().Update(ctx, tenant)
			if err != nil {
				log.Error(err, "Failed to update final tenant status")
				return ctrl.Result{}, err
			}
			err = r.Update(ctx, tenant)
			if err != nil {
				log.Error(err, "Failed to update tenant resource")
				return ctrl.Result{}, err
			}
		}

	} else {
		tenant.Spec.State = "Deleting"
		err = r.Update(ctx, tenant)
		if err != nil {
			log.Error(err, "Failed to update tenant state")
			return ctrl.Result{Requeue: true}, err
		}
		if controllerutil.ContainsFinalizer(tenant, tenantFinalizer) {
			// Run finalization logic for tenantFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			result, err := r.finalizeTenant(ctx, log, tenant)
			if err != nil {
				return result, err
			} else if result.Requeue {
				return result, nil
			}

			// Remove tenantFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(tenant, tenantFinalizer)
			err = r.Update(ctx, tenant)
			if err != nil {
				log.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(tenant, tenantFinalizer) {
		controllerutil.AddFinalizer(tenant, tenantFinalizer)
		err = r.Update(ctx, tenant)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Tenant{}).
		Complete(r)
}

func (r *TenantReconciler) BuildRootTreeStructure(mgr ctrl.Manager) error {
	namespaces := []string{"tenants", "slurm-operator", "tapms-operator"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, ns := range namespaces {
		_, err := v1.CreateHierarchyConfigForNs(ctx, r.Log, r.Client, "multi-tenancy", ns)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to creating hierarchy for %s namespace", ns))
			return err
		}
	}
	_, err := v1.PropagateSecret(ctx, r.Log, "default", "slurm-operator", "wlm-s3-credentials")
	if err != nil {
		return err
	}

	//
	// Since this is called before cache is enabled, need to
	// create/initialize a client.
	//
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	newClient, err := client.New(cfg, client.Options{Scheme: mgr.GetScheme()})
	if err != nil {
		return err
	}
	_, err = v1.AddObjectPropagation(ctx, r.Log, newClient)
	if err != nil {
		return err
	}

	return nil
}

func (r *TenantReconciler) finalizeTenant(ctx context.Context, log logr.Logger, t *v1.Tenant) (ctrl.Result, error) {
	//
	// First delete the child namespaces/anchors
	//
	result, err := v1.DeleteChildNamespaces(ctx, log, r.Client, t, t.Spec.ChildNamespaces)
	if err != nil {
		return result, err
	}

	//
	// Now delete the parent namespace/anchor
	//
	log.Info("Deleting parent namespace: " + t.Spec.TenantName)
	anchor := v1.SubNSAnchorForTenant("tenants", t.Spec.TenantName)
	err = r.Client.Delete(ctx, anchor)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Parent namespace already deleted: " + t.Spec.TenantName)
		} else if k8serrors.IsForbidden(err) {
			log.Info("Requeuing deletion of " + t.Spec.TenantName + ", not ready for deletion yet")
			return ctrl.Result{Requeue: true}, nil
		} else {
			log.Error(err, "Failed to delete parent namespace: "+t.Spec.TenantName)
			return ctrl.Result{}, err
		}
	}

	for _, resource := range t.Spec.TenantResources {
		if len(resource.HsmPartitionName) > 0 {
			log.Info(fmt.Sprintf("Deleting HSM partition %s for tenant %s and resource type %s", resource.HsmPartitionName, t.Spec.TenantName, resource.Type))
			result, err = v1.DeleteHSMPartition(ctx, log, resource.HsmPartitionName)
			if err != nil {
				log.Error(err, "Failed to delete HSM partition")
				return result, err
			}
		}
	}

	for _, resource := range t.Spec.TenantResources {
		if len(resource.HsmGroupLabel) > 0 {
			log.Info(fmt.Sprintf("Deleting HSM group %s for tenant %s and resource type %s", resource.HsmGroupLabel, t.Spec.TenantName, resource.Type))
			result, err = v1.DeleteHSMGroup(ctx, log, resource.HsmGroupLabel)
			if err != nil {
				log.Error(err, "Failed to delete HSM group")
				return result, err
			}
		}
	}

	log.Info("Deleting Keycloak group for: " + t.Spec.TenantName)
	result, err = v1.DeleteKeycloakGroup(ctx, log, t)
	if err != nil {
		log.Error(err, "Failed to delete Keycloak group")
		return result, err
	}

	return ctrl.Result{}, nil
}
