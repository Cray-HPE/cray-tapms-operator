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

package v1alpha1

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"

	uuid "github.com/google/uuid"
)

// TODO: The definitions below will need to be configurable by the site. For now,
// they are tracked below.

// The predefined TAPMS Vault role. This defines Vault actions that the client based on
// the K8s service account will be allowed to perform.
var tapms_vault_role = "tapms-operator"

// The K8s service account token used to perform Vault authentication.
var k8s_service_account_token_path = "/run/secrets/kubernetes.io/serviceaccount/token"

// The tanant Vault transit engine name prefix.
var tapms_transit_prefix = "cray-tenant-"

// The default name of the first Vault transit key used when a new transit engine is created.
var tapms_transit_default_key_name = "key1"

// The default transit engine key algorithm.
// See https://developer.hashicorp.com/vault/api-docs/secret/transit
var tapms_transit_default_key_type = "rsa-2048"

func engineKeyValuePairs(m map[string]interface{}) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}

// Create the tenant Vault transit engine
func CreateVaultTransit(ctx context.Context, log logr.Logger, t *Tenant) (ctrl.Result, error) {
	fmt.Println("CreateVaultTransit called")
	log.Info(fmt.Sprintf("CreateVaultTransit called for tenant (%s)", t.Spec.TenantName))

	// Get Vault client
	client, err := GetVaultClient(log)
	if err != nil {
		// Failed to get Vault token
		return ctrl.Result{}, err
	}

	// TODO: Lookup the tenant transit name. Where is that stored?
	// It will be of the form cray-tenant-$uuid
	// If not found in our record, just create it. If found, check to see if it exists in Vault
	// and if not, create it.

	// Check for the transit engine. Create if it does not exist.
	engine_name := fmt.Sprintf("%s%s", tapms_transit_prefix t.Spec.TenantName, uuid.NewString())
	// This should be the same as calling "vault read sys/mounts/cray-tenant-<name>"
	// The mount point is used when creating or retrieving the transit engine.
	transit_mount_point := fmt.Sprintf("sys/mounts/%s", engine_name)
	log.Info(fmt.Sprintf("Looking for Vault transit engine by name (%s) at location (%s).", engine_name, transit_mount_point))

	// Create the transit engine if not found
	engine, err := client.Logical().Read(transit_mount_point)
	if err != nil {
		// This actually returns an error when the engine is not found.
		// When this happens, the error message will contain:
		// "No secret engine mount at cray-tenant-<name>"
		// TBD: Is there a better way to detect this? It might be cleaner to use the rest api but the
		// return code is 400 with data {"errors":["No secret engine mount at cray-tenant-missing/"]}, so
		// the errors string will still need to be checked.

		if strings.Contains(err.Error(), fmt.Sprintf("No secret engine mount at %s", engine_name)) {
			log.Info(fmt.Sprintf("Did not find existing transit engine (%s).", engine_name))

			log.Info(fmt.Sprintf("Creating new transit engine now. Name (%s)", engine_name))
			engine_info := map[string]interface{}{
				"type":        "transit",
				"description": fmt.Sprintf("%s", t.Spec.TenantName),
			}
			_, err := client.Logical().Write(transit_mount_point, engine_info)
			if err != nil {
				// We had an issue creating the engine.
				return ctrl.Result{}, err
			}

			// TODO: save the transit engine name into k8s. Where? How? 
		} else {
			// We had some other type of error.
			return ctrl.Result{}, err
		}
	} else {
		log.Info(fmt.Sprintf("Found existing Vault transit engine by name (%s).", engine_name))
		log.Info(fmt.Sprintf("DEBUG engine.Data (%s).", engineKeyValuePairs(engine.Data)))
	}

	// Check that we can find the engine. It should exist or have been created at this point.
	_, err = client.Logical().Read(transit_mount_point)
	if err != nil {
		return ctrl.Result{}, err
	}
	//if engine == nil || engine.Data["data"] == nil {
	// What data, if any, are we expecting here? Is this a good check?
	//	return ctrl.Result{}, fmt.Errorf("Nil transit engine or data after creation. Engine: (%s). Unable to continue.", engine_name)
	//}

	// Check that we have the expected default encryption key. Create that if not found.
	log.Info(fmt.Sprintf("Checking for the key %s in the transit engine %s", tapms_transit_default_key_name, engine_name))

	// This should be the same as calling "vault read cray-tenant-<name>/keys/<key-name>"
	// The mount point is used when creating or retrieving the transit engine.
	transit_key_mount_point := fmt.Sprintf("%s/keys/%s", engine_name, tapms_transit_default_key_name)

	// Check that we can find the engine. It should exist at this point.
	transit_key_data, err := client.Logical().Read(transit_key_mount_point)
	if err != nil {
		//TODO: check the behavior here. Is an error thrown when the key does not exist? If so, detect and handle.
		return ctrl.Result{}, err
	}

	if transit_key_data == nil {
		log.Info(fmt.Sprintf("Creating transit key for tenant (%s)", t.Spec.TenantName))
		// TODO: Create the default transit key if it does not exist.
		key_info := map[string]interface{}{
			"type": tapms_transit_default_key_type,
		}
		_, err = client.Logical().Write(transit_key_mount_point, key_info)
		if err != nil {
			return ctrl.Result{}, err
		}
		//if key == nil || key.Data["data"] == nil {
		// What data, if any, are we expecting here? Is this a good check?
		//	return ctrl.Result{}, fmt.Errorf("Nil transit key or data. Engine: (%s), key: (%s). Unable to continue.", engine_name, tapms_transit_default_key_name)
		//}
	} else {
		log.Info(fmt.Sprintf("Found existing transit key for tenant (%s)", t.Spec.TenantName))
	}

	log.Info(fmt.Sprintf("CreateVaultTransit complete for tenant (%s)", t.Spec.TenantName))

	// On success
	return ctrl.Result{}, nil
}

// Delete the tenant Vault transit engine
func DeleteVaultTransit(ctx context.Context, log logr.Logger, t *Tenant) (ctrl.Result, error) {
	log.Info(fmt.Sprintf("DeleteVaultTransit called for tenant (%s)", t.Spec.TenantName))

	// Get Vault client
	client, err := GetVaultClient(log)
	if err != nil {
		// Failed to get Vault token
		return ctrl.Result{}, err
	}
	// Check for the transit engine. Create if it does not exist.
	engine_name := fmt.Sprintf("%s%s", tapms_transit_prefix, t.Spec.TenantName)
	transit_mount_point := fmt.Sprintf("sys/mounts/%s", engine_name)
	log.Info(fmt.Sprintf("Looking for Vault transit engine by name (%s) at location (%s).", engine_name, transit_mount_point))

	// Delete the transit engine if found.
	_, err = client.Logical().Read(transit_mount_point)
	if err == nil {
		log.Info(fmt.Sprintf("Deleting Vault transit by name (%s) at the location (%s).", engine_name, transit_mount_point))
		_, err := client.Logical().Delete(transit_mount_point)
		if err != nil {
			// We had an issue creating the engine.
			return ctrl.Result{}, err
		}
	} else {
		log.Info(fmt.Sprintf("A Vault transit by name (%s) was not found at the location (%s).", engine_name, transit_mount_point))
	}

	log.Info(fmt.Sprintf("DeleteVaultTransit complete for tenant (%s)", t.Spec.TenantName))

	// On success
	return ctrl.Result{}, nil
}

// Get Vault token
func GetVaultClient(log logr.Logger) (client *vault.Client, err error) {
	// See https://github.com/hashicorp/vault-examples/blob/main/examples/auth-methods/kubernetes/go/example.go

	config := vault.DefaultConfig() // modify for more granular configuration

	client, err = vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	k8sAuth, err := auth.NewKubernetesAuth(
		tapms_vault_role,
		auth.WithServiceAccountTokenPath(k8s_service_account_token_path),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.Background(), k8sAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to log in with Kubernetes auth: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return client, nil
}
