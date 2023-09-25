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
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	apiGateway = getEnvVal("API_GATEWAY", "api-gw-service-nmn.local")
	serverPort = getEnvVal("SERVER_PORT", "80")
)

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

func GetServerPort() string {
	return ":" + serverPort
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

func TenantIsUpdated(tenant *Tenant) bool {
	var isUpdated = false
	if !reflect.DeepEqual(tenant.Status.ChildNamespaces, TranslateSpecNamespacesForStatus(tenant.Spec.TenantName, tenant.Spec.ChildNamespaces)) {
		isUpdated = true
	}

	if !reflect.DeepEqual(tenant.Status.TenantResources, tenant.Spec.TenantResources) {
		isUpdated = true
	}

	if !reflect.DeepEqual(tenant.Status.TenantHooks, tenant.Spec.TenantHooks) {
		isUpdated = true
	}

	return isUpdated
}

func GetChildNamespaceName(tenantName string, specChildNamespace string) string {
	return fmt.Sprintf("%s-%s", tenantName, specChildNamespace)
}

func TranslateStatusNamespace(statusNamespace string) string {
	var parts = strings.Split(statusNamespace, "-")
	if len(parts) < 3 {
		//
		// Already w/out the tenant name prefix
		//
		return statusNamespace
	}

	return parts[2]
}

func TranslateStatusNamespacesForSpec(statusNamespaces []string) []string {
	var specNamespaces []string = statusNamespaces
	size := len(statusNamespaces)
	for i := 0; i < size; i++ {
		specNamespaces[i] = TranslateStatusNamespace(specNamespaces[i])
	}

	return specNamespaces
}

func TranslateSpecNamespacesForStatus(tenantName string, specNamespaces []string) []string {
	statusNamespaces := make([]string, 0)
	for _, specNamespace := range specNamespaces {
		statusNamespaces = append(statusNamespaces, GetChildNamespaceName(tenantName, specNamespace))
	}

	return statusNamespaces
}

func CreateQueryParms(name string, values []string) string {
	var urlParameters []string
	for _, value := range values {
		urlParameters = append(urlParameters, fmt.Sprintf("%s=%s", name, value))
	}
	return strings.Join(urlParameters, "&")
}

func Contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func HasIntersection(slice1 []string, slice2 []string) bool {
	// Create an empty slice to store the intersection elements.
	intersection := make([]string, 0)

	// Iterate over the first slice.
	for _, s1 := range slice1 {
		// Check if the current element is in the second slice.
		for _, s2 := range slice2 {
			if s1 == s2 {
				// If it is, add it to the intersection slice.
				intersection = append(intersection, s1)
				break
			}
		}
	}

	// If the intersection slice is not empty, then there is an intersection.
	return len(intersection) > 0
}
