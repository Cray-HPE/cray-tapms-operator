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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
)

type KeycloakGroup struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
	Id   string `json:"id,omitempty"`
}

func listKeycloakGroups(ctx context.Context, log logr.Logger, t *Tenant) (ctrl.Result, []KeycloakGroup, error) {

	result, token, err := GetToken(ctx, log, true)
	if err != nil {
		return result, nil, err
	}

	keycloakUrl := fmt.Sprintf("%s/admin/realms/shasta/groups", getKeycloakBase())

	result, keycloakGroupBytes, err := buildKeycloakGroupPayload(log, t)
	if err != nil {
		return result, nil, err
	}

	req, err := http.NewRequest(http.MethodGet, keycloakUrl, bytes.NewBuffer(keycloakGroupBytes))
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	HTTPClient := NewHttpClient()
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Info("Created Keycloak group: " + getKeycloakGroupName(t.Spec.TenantName))
		return ctrl.Result{}, nil, errors.New("keycloak returned a non-200 response listing groups")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var keycloakGroupList []KeycloakGroup
	err = json.Unmarshal(body, &keycloakGroupList)
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	return ctrl.Result{}, keycloakGroupList, nil

}

func UpdateKeycloakGroup(ctx context.Context, log logr.Logger, t *Tenant) (ctrl.Result, error) {

	result, groupList, err := listKeycloakGroups(ctx, log, t)
	if err != nil {
		return result, err
	}

	for _, group := range groupList {
		if group.Name == getKeycloakGroupName(t.Spec.TenantName) {
			log.Info("Keycloak group already exists: " + getKeycloakGroupName(t.Spec.TenantName))
			return ctrl.Result{}, nil
		}
	}

	result, token, err := GetToken(ctx, log, true)
	if err != nil {
		return result, err
	}
	keycloakUrl := fmt.Sprintf("%s/admin/realms/shasta/groups", getKeycloakBase())

	result, keycloakGroupBytes, err := buildKeycloakGroupPayload(log, t)
	if err != nil {
		return result, err
	}

	req, err := http.NewRequest(http.MethodPost, keycloakUrl, bytes.NewBuffer(keycloakGroupBytes))
	if err != nil {
		return ctrl.Result{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	HTTPClient := NewHttpClient()
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Info("Created Keycloak group: " + getKeycloakGroupName(t.Spec.TenantName))
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, errors.New("keycloak returned a non-200 response creating/updating group")
}

func DeleteKeycloakGroup(ctx context.Context, log logr.Logger, t *Tenant) (ctrl.Result, error) {

	result, groupList, err := listKeycloakGroups(ctx, log, t)
	if err != nil {
		return result, err
	}

	var groupId string
	for _, group := range groupList {
		if group.Name == getKeycloakGroupName(t.Spec.TenantName) {
			groupId = group.Id
			break
		}
	}

	if len(groupId) <= 0 {
		log.Info("Keycloak group already deleted: " + getKeycloakGroupName(t.Spec.TenantName))
		return ctrl.Result{}, nil
	}

	result, token, err := GetToken(ctx, log, true)
	if err != nil {
		return result, err
	}

	keycloakUrl := fmt.Sprintf("%s/admin/realms/shasta/groups/%s", getKeycloakBase(), groupId)

	req, err := http.NewRequest(http.MethodDelete, keycloakUrl, nil)
	if err != nil {
		return ctrl.Result{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	HTTPClient := NewHttpClient()
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Info("Deleted Keycloak group: " + getKeycloakGroupName(t.Spec.TenantName))
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, errors.New("keycloak returned a non-200 response deleting group")
}

func GetToken(ctx context.Context, log logr.Logger, masterAuth bool) (ctrl.Result, string, error) {

	var data url.Values
	var keycloakUrl string
	var res reconcile.Result
	var err error
	if masterAuth {
		keycloakUrl = fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", getKeycloakBase())
		res, data, err = getMasterTokenUrlValues(ctx)

	} else {
		keycloakUrl = fmt.Sprintf("%s/realms/shasta/protocol/openid-connect/token", getClusterKeycloakBase())
		res, data, err = getTokenUrlValues(ctx)
	}
	if err != nil {
		return res, "", err
	}
	req, err := http.NewRequest(http.MethodPost, keycloakUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return ctrl.Result{}, "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	HTTPClient := NewHttpClient()

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return ctrl.Result{}, "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Error(err, fmt.Sprintf("Failed to get token from keycloak, http code: %d", resp.StatusCode))
		return ctrl.Result{Requeue: true}, "", err
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	return ctrl.Result{}, result["access_token"], nil
}

func buildKeycloakGroupPayload(log logr.Logger, t *Tenant) (ctrl.Result, []byte, error) {

	keycloakGroup := KeycloakGroup{}
	keycloakGroup.Name = getKeycloakGroupName(t.Spec.TenantName)
	keycloakGroup.Path = "/" + keycloakGroup.Name
	keycloakGroupBytes, err := json.Marshal(keycloakGroup)
	if err != nil {
		return ctrl.Result{}, nil, err
	}
	return ctrl.Result{}, keycloakGroupBytes, err
}

func getMasterTokenUrlValues(ctx context.Context) (ctrl.Result, url.Values, error) {
	client, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	foundSecret := &corev1.Secret{}
	err = client.Get(ctx, types.NamespacedName{Name: "keycloak-master-admin-auth", Namespace: "services"}, foundSecret)
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	result, clientId, err := decodeSecretValue(foundSecret.Data, "client-id")
	if err != nil {
		return result, nil, err
	}

	result, username, err := decodeSecretValue(foundSecret.Data, "user")
	if err != nil {
		return result, nil, err
	}

	result, password, err := decodeSecretValue(foundSecret.Data, "password")
	if err != nil {
		return result, nil, err
	}

	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("grant_type", "password")
	data.Set("username", username)
	data.Set("password", password)
	return ctrl.Result{}, data, nil
}

func getTokenUrlValues(ctx context.Context) (ctrl.Result, url.Values, error) {
	client, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	foundSecret := &corev1.Secret{}
	err = client.Get(ctx, types.NamespacedName{Name: "admin-client-auth", Namespace: "default"}, foundSecret)
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	result, clientSecret, err := decodeSecretValue(foundSecret.Data, "client-secret")
	if err != nil {
		return result, nil, err
	}
	result, clientId, err := decodeSecretValue(foundSecret.Data, "client-id")
	if err != nil {
		return result, nil, err
	}

	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("grant_type", "client_credentials")
	data.Set("client_secret", clientSecret)
	return ctrl.Result{}, data, nil
}

func decodeSecretValue(data map[string][]byte, key string) (ctrl.Result, string, error) {
	encStr := base64.StdEncoding.EncodeToString([]byte(data[key]))
	decStr, err := base64.StdEncoding.DecodeString(encStr)
	if err != nil {
		return ctrl.Result{}, "", err
	}
	return ctrl.Result{}, string(decStr), nil
}

func getKeycloakGroupName(tenantName string) string {
	return tenantName + "-tenant-admin"
}

func getKeycloakBase() string {
	return "http://keycloak.services:8080/keycloak"
}

func getClusterKeycloakBase() string {
	return fmt.Sprintf("https://%s/keycloak", GetApiGateway())
}
