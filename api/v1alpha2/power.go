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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

type XnamePowerState struct {
	Xname                     string   `json:"xname,omitempty"`
	PowerState                string   `json:"powerstate,omitempty"`
	ManagementState           string   `json:"managementstate,omitempty"`
	Error                     string   `json:"error,omitempty"`
	LastUpdated               string   `json:"lastupdated,omitempty"`
	SupportedPowerTransitions []string `json:"supportedpowertransitions,omitempty"`
}

type PowerStatus struct {
	Status []XnamePowerState `json:"status,omitempty"`
}

type PowerTransitionLocation struct {
	Xname     string `json:"xname,omitempty"`
	DeputyKey string `json:"deputykey,omitempty"`
}
type PowerTransitionStartOutput struct {
	TransitionID string `json:"transitionid,omitempty"`
	Operation    string `json:"operation,omitempty"`
}

type PowerTransitionRequest struct {
	Operation           string                    `json:"operation,omitempty"`
	TaskDeadlineMinutes int                       `json:"taskdeadlineminutes"`
	Location            []PowerTransitionLocation `json:"location,omitempty"`
}

func EnsurePowerState(ctx context.Context, log logr.Logger, t *Tenant) (ctrl.Result, error) {
	xnamesToPowerOff := make([]string, 0)
	//
	// This loop handles case where a resource group is removed.
	//
	for _, statResource := range t.Status.TenantResources {
		haveSpec := false
		for _, resource := range t.Spec.TenantResources {
			if statResource.Type == resource.Type {
				haveSpec = true
			}
		}
		if !haveSpec {
			xnamesToPowerOff = append(xnamesToPowerOff, statResource.Xnames...)
		}
	}

	//
	// This loop handles case where a resource group is added.
	//
	for _, specResource := range t.Spec.TenantResources {
		haveStatus := false
		for _, resource := range t.Status.TenantResources {
			if specResource.Type == resource.Type {
				haveStatus = true
			}
		}
		if !haveStatus {
			xnamesToPowerOff = append(xnamesToPowerOff, specResource.Xnames...)
		}
	}

	for _, specResource := range t.Spec.TenantResources {
		for _, statResource := range t.Status.TenantResources {
			if statResource.Type == specResource.Type {
				log.Info(fmt.Sprintf("Checking for members deleted from HSM partition %s", specResource.HsmPartitionName))
				deletedMembers := Difference(statResource.Xnames, specResource.Xnames)
				addedMembers := Difference(specResource.Xnames, statResource.Xnames)
				xnamesToPowerOff = append(xnamesToPowerOff, deletedMembers...)
				xnamesToPowerOff = append(xnamesToPowerOff, addedMembers...)
			}
		}
	}
	result, err := ensureXnamesOff(ctx, log, xnamesToPowerOff)
	if err != nil {
		log.Error(err, "Failed to power off xname(s)")
		return result, err
	}

	return ctrl.Result{}, nil
}

func ensureXnamesOff(ctx context.Context, log logr.Logger, xnames []string) (ctrl.Result, error) {
	if len(xnames) == 0 {
		return ctrl.Result{}, nil
	}

	result, powerStatus, err := getPowerStatus(ctx, log, xnames)
	if err != nil {
		log.Error(err, "Failed to check power status")
		return result, err
	}

	xnamesToPowerOff := make([]string, 0)
	for _, state := range powerStatus.Status {
		if !Contains(xnames, state.Xname) {
			continue
		}
		log.Info(fmt.Sprintf("Current power status %s: %v", state.Xname, state.PowerState))
		if state.PowerState != "off" {
			log.Info(fmt.Sprintf("Requesting power off for %s", state.Xname))
			xnamesToPowerOff = append(xnamesToPowerOff, state.Xname)
		}
	}
	result, err = powerOffXnames(ctx, log, xnamesToPowerOff)
	if err != nil {
		log.Error(err, "Failed to power off xnames")
		return result, err
	}

	return ctrl.Result{}, nil
}

func powerOffXnames(ctx context.Context, log logr.Logger, xnames []string) (ctrl.Result, error) {

	if len(xnames) == 0 {
		return ctrl.Result{}, nil
	}

	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}
	result, pRequestBytes, err := buildPowerTransitionReqPayload(log, xnames)
	if err != nil {
		return result, err
	}

	powerUrl := fmt.Sprintf("https://%s/apis/power-control/v1/transitions", GetApiGateway())
	req, err := http.NewRequest(http.MethodPost, powerUrl, bytes.NewBuffer(pRequestBytes))
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

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ctrl.Result{}, errors.New("PCS returned a non-200 response requesting power transition")
	}
	var powerTransitionStartOutput PowerTransitionStartOutput
	err = json.Unmarshal(body, &powerTransitionStartOutput)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("Power Transition ID: %s for xnames %v", powerTransitionStartOutput.TransitionID, xnames))
	return ctrl.Result{}, nil
}

func getPowerStatus(ctx context.Context, log logr.Logger, xnames []string) (ctrl.Result, *PowerStatus, error) {

	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, nil, err
	}

	queryParms := CreateQueryParms("xname", xnames)
	powerUrl := fmt.Sprintf("https://%s/apis/power-control/v1/power-status?%s", GetApiGateway(), queryParms)
	req, err := http.NewRequest(http.MethodGet, powerUrl, nil)
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

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ctrl.Result{}, nil, errors.New("PCS returned a non-200 response getting power status")
	}

	var status PowerStatus
	err = json.Unmarshal(body, &status)
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	return ctrl.Result{}, &status, nil
}

func buildPowerTransitionReqPayload(log logr.Logger, xnames []string) (ctrl.Result, []byte, error) {

	pRequest := PowerTransitionRequest{}
	pRequest.Operation = "off"
	pRequest.TaskDeadlineMinutes = 0
	for _, xname := range xnames {
		location := PowerTransitionLocation{}
		location.Xname = xname
		pRequest.Location = append(pRequest.Location, location)
	}

	pRequestBytes, err := json.Marshal(pRequest)
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	return ctrl.Result{}, pRequestBytes, err
}
