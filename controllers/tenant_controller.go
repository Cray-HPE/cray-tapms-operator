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

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	api "sigs.k8s.io/hierarchical-namespaces/api/v1alpha2"

	"github.com/Cray-HPE/cray-tapms-operator/api/v1alpha1"
	tapmshpecomv1alpha1 "github.com/Cray-HPE/cray-tapms-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const tenantFinalizer = "tapms.hpe.com/finalizer"

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *TenantReconciler) subNSAnchorForTenant(parentObj metav1.Object, parentNs string, childNs string) *api.SubnamespaceAnchor {

	anchor := &api.SubnamespaceAnchor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      childNs,
			Namespace: parentNs,
		},
	}
	//controllerutil.SetControllerReference(parentObj, anchor, r.Scheme)
	return anchor
}

func (r *TenantReconciler) ensureSubanchorNSExists(log logr.Logger, ctx context.Context, parentObj metav1.Object, parentNs string, childNs string) (*api.SubnamespaceAnchor, ctrl.Result, error) {

	subNsAnchor := &api.SubnamespaceAnchor{}
	log.Info("Checking to see if tenant subanchor exists: " + childNs + " in parent namespace: " + parentNs)
	err := r.Client.Get(ctx, types.NamespacedName{Namespace: parentNs, Name: childNs}, subNsAnchor)

	// subanchor does not exist, try to create it
	if err != nil && k8serrors.IsNotFound(err) {
		log.Info("Tenant subanchor not found, creating: " + childNs)
		subNsAnchor = r.subNSAnchorForTenant(parentObj, parentNs, childNs)
		err = r.Client.Create(ctx, subNsAnchor)
		if err != nil {
			return nil, ctrl.Result{}, err
		}
		return subNsAnchor, ctrl.Result{}, nil
	} else if err != nil {
		return nil, ctrl.Result{}, err
	}

	return subNsAnchor, ctrl.Result{}, nil
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
	tenant := &tapmshpecomv1alpha1.Tenant{}
	err := r.Get(ctx, req.NamespacedName, tenant)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("CR not found (deleted?). Ignoring")
			return ctrl.Result{}, nil // don't reqeueue
		}
	}

	parentSubNsAnchor, result, err := r.ensureSubanchorNSExists(log, ctx, tenant, "tenants", tenant.Spec.TenantName)

	if err != nil {
		return result, err
	}

	isTenantMarkedToBeDeleted := tenant.GetDeletionTimestamp() != nil
	if !isTenantMarkedToBeDeleted {
		if tenant.Spec.ChildNamespaces != nil {
			for _, childNamespace := range tenant.Spec.ChildNamespaces {
				childNs := tenant.Spec.TenantName + "-" + childNamespace
				_, result, err := r.ensureSubanchorNSExists(log, ctx, parentSubNsAnchor, tenant.Spec.TenantName, childNs)
				if err != nil {
					return result, err
				}
			}
		}
	}

	if isTenantMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(tenant, tenantFinalizer) {
			// Run finalization logic for tenantFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeTenant(ctx, log, tenant); err != nil {
				return ctrl.Result{}, err
			}

			// Remove tenantFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(tenant, tenantFinalizer)
			err := r.Update(ctx, tenant)
			if err != nil {
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
		For(&tapmshpecomv1alpha1.Tenant{}).
		//		Owns(&api.SubnamespaceAnchor{}).
		Complete(r)
}

func (r *TenantReconciler) finalizeTenant(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant) error {
	//
	// First delete the child namespaces/anchors
	//
	for _, childNamespace := range t.Spec.ChildNamespaces {
		childNs := t.Spec.TenantName + "-" + childNamespace
		log.Info("Deleting child namespace: " + childNs)
		anchor := r.subNSAnchorForTenant(nil, t.Spec.TenantName, childNs)
		err := r.Client.Delete(ctx, anchor)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Info("Child namespace already deleted: " + childNs)
			} else {
				log.Error(err, "Failed to delete child namespace: "+childNs)
				return err
			}
		}
	}

	//
	// Now delete the parent namespace/anchor
	//
	log.Info("Deleting parent namespace: " + t.Spec.TenantName)
	anchor := r.subNSAnchorForTenant(nil, "tenants", t.Spec.TenantName)
	err := r.Client.Delete(ctx, anchor)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Parent namespace already deleted: " + t.Spec.TenantName)
		} else {
			log.Error(err, "Failed to delete parent namespace: "+t.Spec.TenantName)
			return err
		}
	}

	return nil
}
