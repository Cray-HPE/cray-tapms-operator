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

package lib

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"reflect"

	"github.com/Cray-HPE/cray-tapms-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var apiGateway = getEnvVal("API_GATEWAY", "api-gw-service-nmn.local")

func NewHttpClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	httpClient := &http.Client{Transport: transport}

	return httpClient
}

func GetApiGateway() string {
	return apiGateway
}

func Difference(a, b []string) (diff []string) {
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}

func PropagateSecret(ctx context.Context, log logr.Logger, fromNs string, toNs string, secretName string) (ctrl.Result, error) {
	log.Info(fmt.Sprintf("Ensuring secret %s exists in %s namespace", secretName, toNs))
	client, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		return ctrl.Result{}, err
	}
	toSecret := &corev1.Secret{}
	err = client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: toNs}, toSecret)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	if toSecret.Name == secretName {
		log.Info(fmt.Sprintf("Secret %s already exists in %s namespace", secretName, toNs))
		return ctrl.Result{}, nil
	}

	fromSecret := &corev1.Secret{}
	err = client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: fromNs}, fromSecret)
	if err != nil {
		return ctrl.Result{}, err
	}
	newSecret := &corev1.Secret{}
	newSecret.Namespace = toNs
	newSecret.Name = fromSecret.Name
	newSecret.ObjectMeta.Namespace = toNs
	newSecret.Data = fromSecret.Data
	newAnnotations := map[string]string{
		"namespace": toNs,
		"name":      fromSecret.Name,
	}
	newSecret.Annotations = newAnnotations

	err = client.Create(ctx, newSecret)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Secret %s created in %s namespace", secretName, toNs))

	return ctrl.Result{}, nil
}

func getEnvVal(envVar, defVal string) string {
	if e, ok := os.LookupEnv(envVar); ok {
		return e
	}
	return defVal
}

func TenantIsUpdated(tenant *v1alpha1.Tenant) bool {
	var isUpdated = false
	if !reflect.DeepEqual(tenant.Status.Xnames, tenant.Spec.TenantResource.Xnames) {
		isUpdated = true
	}
	if !reflect.DeepEqual(tenant.Status.ChildNamespaces, tenant.Spec.ChildNamespaces) {
		isUpdated = true
	}
	if tenant.Status.HsmPartitionName != tenant.Spec.TenantResource.HsmPartitionName {
		isUpdated = true
	}
	return isUpdated
}
