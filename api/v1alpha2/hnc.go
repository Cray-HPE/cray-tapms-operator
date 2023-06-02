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

package v1alpha2

import (
	"context"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	api "sigs.k8s.io/hierarchical-namespaces/api/v1alpha2"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeleteChildNamespaces(ctx context.Context, log logr.Logger, client client.Client, t *Tenant, childNamespaces []string) (ctrl.Result, error) {
	for _, childNamespace := range childNamespaces {
		childNs := GetChildNamespaceName(t.Spec.TenantName, childNamespace)
		log.Info("Deleted child namespace: " + childNs)
		anchor := SubNSAnchorForTenant(t.Spec.TenantName, childNs)
		err := client.Delete(ctx, anchor)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Info("Child namespace already deleted: " + childNs)
			} else {
				log.Error(err, "Failed to delete child namespace: "+childNs)
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func SubNSAnchorForTenant(parentNs string, childNs string) *api.SubnamespaceAnchor {

	anchor := &api.SubnamespaceAnchor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      childNs,
			Namespace: parentNs,
		},
	}
	return anchor
}

func HierarchyConfigForTenant(parentNs string, childNs string) *api.HierarchyConfiguration {

	hierarchyConfigSpec := &api.HierarchyConfigurationSpec{
		Parent: parentNs,
	}
	hierarchy := &api.HierarchyConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hierarchy",
			Namespace: childNs,
		},
		Spec: *hierarchyConfigSpec,
	}
	return hierarchy
}

func CreateSubanchorNs(ctx context.Context, log logr.Logger, client client.Client, parentNs string, childNs string) (ctrl.Result, error) {
	subNsAnchor := SubNSAnchorForTenant(parentNs, childNs)
	err := client.Create(ctx, subNsAnchor)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			log.Info("Subanchor: " + childNs + " in parent namespace: " + parentNs + " already exists")
			return ctrl.Result{}, nil
		} else if k8serrors.IsNotFound(err) {
			//
			// It can take the hnc-manager a bit to create namespaces,
			// so if we get namespace not found, we'll try again.
			//
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	log.Info("Created subanchor: " + childNs + " in parent namespace: " + parentNs)
	return ctrl.Result{}, nil
}

func CreateHierarchyConfigForNs(ctx context.Context, log logr.Logger, client client.Client, parentNs string, childNs string) (ctrl.Result, error) {
	hierarchyConfig := HierarchyConfigForTenant(parentNs, childNs)
	err := client.Create(ctx, hierarchyConfig)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			log.Info("hierarchyConfig: " + childNs + " in parent namespace: " + parentNs + " already exists")
			return ctrl.Result{}, nil
		} else if k8serrors.IsNotFound(err) {
			//
			// It can take the hnc-manager a bit to create hierarchies,
			// so if we get not found, we'll try again.
			//
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	log.Info("Created hierarchyConfig: " + childNs + " in parent namespace: " + parentNs)
	return ctrl.Result{}, nil
}

func HNCConfiguration(resourceVersion string) *api.HNCConfiguration {
	limitRangeSpec := &api.ResourceSpec{
		Group:    "",
		Resource: "limitrange",
		Mode:     "Propagate",
	}
	resourceQuotaSpec := &api.ResourceSpec{
		Group:    "",
		Resource: "resourcequota",
		Mode:     "Propagate",
	}
	rolesQuotaSpec := &api.ResourceSpec{
		Group:    "",
		Resource: "roles",
		Mode:     "Propagate",
	}
	rolesBindingsSpec := &api.ResourceSpec{
		Group:    "",
		Resource: "rolebindings",
		Mode:     "Propagate",
	}

	hncConfigurationSpec := &api.HNCConfigurationSpec{
		Resources: []api.ResourceSpec{*limitRangeSpec, *resourceQuotaSpec, *rolesQuotaSpec, *rolesBindingsSpec},
	}

	hncConfiguration := &api.HNCConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "config",
			ResourceVersion: resourceVersion,
		},
		Spec: *hncConfigurationSpec,
	}
	return hncConfiguration
}

func AddObjectPropagation(ctx context.Context, log logr.Logger, client client.Client) (ctrl.Result, error) {
	currentConfig := &api.HNCConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "config",
		},
	}
	err := client.Get(ctx, types.NamespacedName{Name: "config", Namespace: "hnc-system"}, currentConfig)
	if err != nil {
		return ctrl.Result{}, err
	}

	hncConfig := HNCConfiguration(currentConfig.ObjectMeta.ResourceVersion)
	err = client.Update(ctx, hncConfig)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("Successful update of resources for propogation: '%v'", hncConfig.Spec.Resources))
	return ctrl.Result{}, nil
}
