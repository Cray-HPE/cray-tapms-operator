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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Cray-HPE/cray-tapms-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

type hsmIds struct {
	Ids []string
}

type hsmPartition struct {
	Name        string
	Description string
	Tags        []string
	Members     hsmIds
}

func UpdateHSMPartition(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant, members []v1alpha1.TenantResource) (ctrl.Result, error) {

	result, token, err := GetToken(ctx, log)
	if err != nil {
		return result, err
	}
	//hsmUrl := "https://api-gw-service-nmn.local/apis/smd/hsm/v2/partitions"
	hsmUrl := "https://api-gateway.vshasta.io/apis/smd/hsm/v2/partitions"
	hsmPartition := hsmPartition{}
	hsmPartition.Name = t.Name
	for _, member := range members {
		for _, xname := range member.Xnames {
			hsmPartition.Members.Ids = append(hsmPartition.Members.Ids, xname)
		}
	}
	hsmPartitionBytes, err := json.Marshal(hsmPartition)
	if err != nil {
		return ctrl.Result{}, err
	}
	req, err := http.NewRequest(http.MethodPost, hsmUrl, bytes.NewBuffer(hsmPartitionBytes))
	if err != nil {
		return ctrl.Result{}, err
	}

	req.Header.Set("Context-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	fmt.Printf("REQ: %+v\n", req)
	HTTPClient := NewHttpClient()
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer resp.Body.Close()

	fmt.Printf("RESP: %+v\n", resp)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("RESP BODY: %+v\n", body)
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, errors.New("HSM returned a non-200 response")
}
