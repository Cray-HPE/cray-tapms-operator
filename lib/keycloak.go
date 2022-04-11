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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/go-logr/logr"
)

func GetToken(ctx context.Context, log logr.Logger) (ctrl.Result, string, error) {
	keycloakUrl := "https://api-gateway.vshasta.io/keycloak/realms/shasta/protocol/openid-connect/token"

	res, data, err := getTokenUrlValues(ctx)
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
