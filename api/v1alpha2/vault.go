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

package v1alpha2

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/google/uuid"
	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
)

// The definitions below may need to be configurable by the site. For now,
// they are tracked below.

// The predefined TAPMS Vault role. This defines Vault actions that the client, based on
// the K8s service account, will be allowed to perform.
var tapms_vault_role = "tapms-operator"

// The K8s service account token used to perform Vault authentication.
var k8s_service_account_token_path = "/var/run/secrets/kubernetes.io/serviceaccount/token"

// The tanant Vault transit engine name prefix.
var tapms_transit_prefix = "cray-tenant-"

// Create the tenant Vault transit engine
func CreateVaultTransit(ctx context.Context, log logr.Logger, t *Tenant) (ctrl.Result, error) {
	fmt.Println("CreateVaultTransit called")
	log.Info(fmt.Sprintf("CreateVaultTransit called for tenant (%s)", t.Spec.TenantName))

	createTransit := t.Spec.TenantKmsResource.Enabled
	log.Info(fmt.Sprintf("enablekms=%t", createTransit))

	if createTransit {
		// Get Vault client
		client, err := GetVaultClient(log)
		if err != nil {
			// Failed to get Vault token
			return ctrl.Result{}, err
		}

		// Get the transit engine key attributes from the specification.
		// See the tenant_types and the generated CRD for defaults if these are not set.
		// See https://developer.hashicorp.com/vault/api-docs/secret/transit#type for possible key types.
		transit_engine_key_name := t.Spec.TenantKmsResource.KeyName
		transit_engine_key_type := t.Spec.TenantKmsResource.KeyType

		// Check if a tenant transit name was previously reocrded in the status.
		// It will be of the form cray-tenant-$uuid.
		engine_name := t.Status.TenantKmsStatus.TransitName
		if engine_name == "" {
			// If not found in the status, generate it now.
			if t.Status.UUID == "" {
				// If we have no tenant UUID, generate a new UUID for the transit engine name.
				engine_name = fmt.Sprintf("%s%s", tapms_transit_prefix, uuid.New().String())
			} else {
				// Include the tenant UUID in the new transit engine name.
				engine_name = fmt.Sprintf("%s%s", tapms_transit_prefix, t.Status.UUID)
			}
		}

		// Check for the transit engine. Create if it does not exist.
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

				// Record the transit engine name.
				// The tenant controller will update the status with this info.
				t.Status.TenantKmsStatus.TransitName = engine_name
			} else {
				// We had some other type of error.
				return ctrl.Result{}, err
			}
		} else {
			log.Info(fmt.Sprintf("Found existing Vault transit engine by name (%s).", engine_name))
		}

		// Check that we can find the engine. It should exist or have been created at this point.
		_, err = client.Logical().Read(transit_mount_point)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Check that we have the expected default encryption key. Create that if not found.
		log.Info(fmt.Sprintf("Checking for the key %s in the transit engine %s", transit_engine_key_name, engine_name))

		// This should be the same as calling "vault read cray-tenant-<name>/keys/<key-name>"
		// The mount point is used when creating or retrieving the transit engine.
		transit_key_mount_point := fmt.Sprintf("%s/keys/%s", engine_name, transit_engine_key_name)

		// Check that we can find the engine. It should exist at this point.
		transit_key_data, err := client.Logical().Read(transit_key_mount_point)
		if err != nil {
			return ctrl.Result{}, err
		}

		if transit_key_data == nil {
			log.Info(fmt.Sprintf("Creating transit key for tenant (%s)", t.Spec.TenantName))
			key_info := map[string]interface{}{
				"type": transit_engine_key_type,
			}
			_, err = client.Logical().Write(transit_key_mount_point, key_info)
			if err != nil {
				return ctrl.Result{}, err
			}

			// Record the transit engine key name and type.
			// The tenant controller will update the status with this info.
			t.Status.TenantKmsStatus.KeyName = transit_engine_key_name
			t.Status.TenantKmsStatus.KeyType = transit_engine_key_type

			// Read the transit key metadata
			transit_key_data, err = client.Logical().Read(transit_key_mount_point)
			if err != nil {
				return ctrl.Result{}, err
			}
			if transit_key_data == nil {
				log.Info(fmt.Sprintf("Nil transit key data for mount point(%s)", transit_key_mount_point))
				return ctrl.Result{}, err
			} else {
				// Display the transit key metadata as json in the k8s tapms status.
				// Note: to see more detail such as the min/max supported encryption version,
				// marshal the entire transit_key_data structure. This will be useful when
				// tenant admins start to work with key rotation. For now, this form will display whatever
				// data is available for "keys". In this form, if someone had performed key
				// rotation in Vault, multiple keys will be listed. It will be up to the tenant
				// admin to know which key to use since any key rotation is outisde of scope of
				// what tapms is responsible for managing.
				jsonStr, err := json.Marshal(transit_key_data.Data["keys"])
				if err != nil {
					fmt.Printf("Error: %s", err.Error())
				} else {
					t.Status.TenantKmsStatus.PublicKey = string(jsonStr)
				}
			}

		} else {
			log.Info(fmt.Sprintf("Found existing transit key for tenant (%s)", t.Spec.TenantName))
			// TODO: Update the k8s status (t.Status.TenantKmsStatus.PublicKey) here again
			// to pick up any new key(s) that could have been created by rotation.
		}
	} else {
		// The case where t.Spec.TenantKmsResource.Enabled=false
		log.Info(fmt.Sprintf("No transit engine was requested for tenant (%s)", t.Spec.TenantName))
	}
	log.Info(fmt.Sprintf("CreateVaultTransit complete for tenant (%s)", t.Spec.TenantName))

	// On success
	return ctrl.Result{}, nil
}

// Delete the tenant Vault transit engine
func DeleteVaultTransit(ctx context.Context, log logr.Logger, t *Tenant) (ctrl.Result, error) {
	log.Info(fmt.Sprintf("DeleteVaultTransit called for tenant (%s)", t.Spec.TenantName))

	// Check if a tenant transit name was previously reocrded in the status.
	// It will be of the form cray-tenant-$uuid.
	engine_name := t.Status.TenantKmsStatus.TransitName
	if engine_name == "" {
		// We have nothing to delete if the transit engine name was not found in the status.
		log.Info(fmt.Sprintf("Did not fine a Vault transit engine for the tenant (%s).", t.Spec.TenantName))
	} else {
		// Get Vault client
		client, err := GetVaultClient(log)
		if err != nil {
			// Failed to get Vault token
			return ctrl.Result{}, err
		}
		// Check for the transit engine. Create if it does not exist.
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
