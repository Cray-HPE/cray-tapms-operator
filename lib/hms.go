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

type HsmMemberId struct {
	Id string
}

type HsmIds struct {
	Ids []string
}

type HsmPartition struct {
	Name        string
	Description string
	Tags        []string
	Members     HsmIds
}

type HsmGroup struct {
	Label       string
	Description string
	Tags        []string
	Members     HsmIds
}

func ListHSMGroups(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant) (ctrl.Result, []HsmGroup, error) {

	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, nil, err
	}
	hsmUrl := fmt.Sprintf("https://%s/apis/smd/hsm/v2/groups", GetApiGateway())
	hsmGroup := HsmGroup{}
	hsmGroupBytes, err := json.Marshal(hsmGroup)
	if err != nil {
		return ctrl.Result{}, nil, err
	}
	req, err := http.NewRequest(http.MethodGet, hsmUrl, bytes.NewBuffer(hsmGroupBytes))
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
		return ctrl.Result{}, nil, errors.New("HSM returned a non-200 response listing groups")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var hsmGroupList []HsmGroup
	err = json.Unmarshal(body, &hsmGroupList)
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	return ctrl.Result{}, hsmGroupList, nil
}

func ListHSMPartitions(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant) (ctrl.Result, []HsmPartition, error) {

	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, nil, err
	}
	hsmUrl := fmt.Sprintf("https://%s/apis/smd/hsm/v2/partitions", GetApiGateway())
	hsmPartition := HsmPartition{}
	hsmPartitionBytes, err := json.Marshal(hsmPartition)
	if err != nil {
		return ctrl.Result{}, nil, err
	}
	req, err := http.NewRequest(http.MethodGet, hsmUrl, bytes.NewBuffer(hsmPartitionBytes))
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
		return ctrl.Result{}, nil, errors.New("HSM returned a non-200 response listing partitions")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var hsmPartitionList []HsmPartition
	err = json.Unmarshal(body, &hsmPartitionList)
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	return ctrl.Result{}, hsmPartitionList, nil
}

func UpdateHSMGroup(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant) (ctrl.Result, error) {

	result, groupList, err := ListHSMGroups(ctx, log, t)
	if err != nil {
		return result, err
	}

	existingGroup := false
	for _, group := range groupList {
		if group.Label == t.Spec.TenantResource.HsmGroupLabel {
			existingGroup = true
			break
		}
	}

	if !existingGroup {
		//
		// create the group
		//
		result, err := createHSMGroup(ctx, log, t)
		if err != nil {
			return result, err
		}
		return ctrl.Result{}, nil
	} else {
		//
		// Check for any changes to update in the group
		//
		log.Info("Updating HSM group")
		deletedMembers := Difference(t.Status.Xnames, t.Spec.TenantResource.Xnames)
		result, err := editHsmGroupMembers(ctx, log, t, deletedMembers, http.MethodDelete)
		if err != nil {
			log.Error(err, "Failed to delete HSM group members")
			return result, err
		}
		addedMembers := Difference(t.Spec.TenantResource.Xnames, t.Status.Xnames)
		result, err = editHsmGroupMembers(ctx, log, t, addedMembers, http.MethodPost)
		if err != nil {
			log.Error(err, "Failed to add HSM group members")
			return result, err
		}
	}
	return ctrl.Result{}, nil
}

func UpdateHSMPartition(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant) (ctrl.Result, error) {

	result, partitionList, err := ListHSMPartitions(ctx, log, t)
	if err != nil {
		return result, err
	}

	existingPartition := false
	for _, partition := range partitionList {
		if partition.Name == t.Spec.TenantResource.HsmPartitionName {
			existingPartition = true
			break
		}
	}

	if !existingPartition {
		//
		// create the partition
		//
		result, err := createHSMPartition(ctx, log, t)
		if err != nil {
			return result, err
		}
		return ctrl.Result{}, nil
	} else {
		//
		// Check for any changes to update in the partition
		//
		log.Info("Updating HSM partition")
		deletedMembers := Difference(t.Status.Xnames, t.Spec.TenantResource.Xnames)
		result, err := editHsmPartitionMembers(ctx, log, t, deletedMembers, http.MethodDelete)
		if err != nil {
			log.Error(err, "Failed to delete HSM partition members")
			return result, err
		}
		addedMembers := Difference(t.Spec.TenantResource.Xnames, t.Status.Xnames)
		result, err = editHsmPartitionMembers(ctx, log, t, addedMembers, http.MethodPost)
		if err != nil {
			log.Error(err, "Failed to add HSM partition members")
			return result, err
		}
	}
	return ctrl.Result{}, nil
}

func editHsmGroupMembers(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant, changedMembers []string, httpMethod string) (ctrl.Result, error) {
	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	for _, member := range changedMembers {

		hsmUrl := ""
		action := ""
		hsmGroupBytes := []byte{}
		if httpMethod == http.MethodPost {
			hsmUrl = fmt.Sprintf("https://%s/apis/smd/hsm/v2/groups/%s/members", GetApiGateway(), t.Spec.TenantResource.HsmGroupLabel)
			action = "adding"
			hsmId := HsmMemberId{}
			hsmId.Id = member
			hsmGroupBytes, err = json.Marshal(hsmId)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if httpMethod == http.MethodDelete {
			hsmUrl = fmt.Sprintf("https://%s/apis/smd/hsm/v2/groups/%s/members/%s", GetApiGateway(), t.Spec.TenantResource.HsmGroupLabel, member)
			action = "removing"
			memberArray := []string{member}
			result, hsmGroupBytes, err = buildHsmGroupPayload(log, t, memberArray)
			if err != nil {
				return result, err
			}
		}

		req, err := http.NewRequest(httpMethod, hsmUrl, bytes.NewBuffer(hsmGroupBytes))
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

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			if resp.StatusCode == 404 {
				log.Info(fmt.Sprintf("HSM member %s already deleted from group %s", member, t.Spec.TenantResource.HsmGroupLabel))
			} else {
				return ctrl.Result{}, fmt.Errorf("HSM returned a non-200 response %s member %s for group %s", action, member, t.Spec.TenantResource.HsmGroupLabel)
			}
		}
	}

	return ctrl.Result{}, nil
}

func editHsmPartitionMembers(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant, changedMembers []string, httpMethod string) (ctrl.Result, error) {
	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	for _, member := range changedMembers {

		hsmUrl := ""
		action := ""
		hsmPartitionBytes := []byte{}
		if httpMethod == http.MethodPost {
			hsmUrl = fmt.Sprintf("https://%s/apis/smd/hsm/v2/partitions/%s/members", GetApiGateway(), t.Spec.TenantResource.HsmPartitionName)
			action = "adding"
			hsmId := HsmMemberId{}
			hsmId.Id = member
			hsmPartitionBytes, err = json.Marshal(hsmId)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if httpMethod == http.MethodDelete {
			hsmUrl = fmt.Sprintf("https://%s/apis/smd/hsm/v2/partitions/%s/members/%s", GetApiGateway(), t.Spec.TenantResource.HsmPartitionName, member)
			action = "removing"
			memberArray := []string{member}
			result, hsmPartitionBytes, err = buildHsmPartitionPayload(log, t, memberArray)
			if err != nil {
				return result, err
			}
		}

		req, err := http.NewRequest(httpMethod, hsmUrl, bytes.NewBuffer(hsmPartitionBytes))
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

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			if resp.StatusCode == 404 {
				log.Info(fmt.Sprintf("HSM member %s already deleted from partition %s", member, t.Spec.TenantResource.HsmPartitionName))
			} else {
				return ctrl.Result{}, fmt.Errorf("HSM returned a non-200 response %s member %s for partition %s", action, member, t.Spec.TenantResource.HsmPartitionName)
			}
		}
	}

	return ctrl.Result{}, nil
}

func buildHsmPartitionPayload(log logr.Logger, t *v1alpha1.Tenant, xnames []string) (ctrl.Result, []byte, error) {

	hsmPartition := HsmPartition{}
	hsmPartition.Name = t.Spec.TenantResource.HsmPartitionName
	hsmPartition.Tags = append(hsmPartition.Tags, t.Name)
	for _, xname := range xnames {
		hsmPartition.Members.Ids = append(hsmPartition.Members.Ids, xname)
	}
	hsmPartitionBytes, err := json.Marshal(hsmPartition)
	if err != nil {
		return ctrl.Result{}, nil, err
	}
	return ctrl.Result{}, hsmPartitionBytes, err
}

func buildHsmGroupPayload(log logr.Logger, t *v1alpha1.Tenant, xnames []string) (ctrl.Result, []byte, error) {

	hsmGroup := HsmGroup{}
	hsmGroup.Label = t.Spec.TenantResource.HsmGroupLabel
	hsmGroup.Tags = append(hsmGroup.Tags, t.Name)
	for _, xname := range xnames {
		hsmGroup.Members.Ids = append(hsmGroup.Members.Ids, xname)
	}
	hsmGroupBytes, err := json.Marshal(hsmGroup)
	if err != nil {
		return ctrl.Result{}, nil, err
	}
	return ctrl.Result{}, hsmGroupBytes, err
}

func createHSMGroup(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant) (ctrl.Result, error) {
	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	hsmUrl := fmt.Sprintf("https://%s/apis/smd/hsm/v2/groups", GetApiGateway())
	result, hsmGroupBytes, err := buildHsmGroupPayload(log, t, t.Spec.TenantResource.Xnames)
	if err != nil {
		return result, err
	}

	req, err := http.NewRequest(http.MethodPost, hsmUrl, bytes.NewBuffer(hsmGroupBytes))
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
		log.Info("Created HSM group: " + t.Spec.TenantResource.HsmGroupLabel)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, errors.New("HSM returned a non-200 response creating group")
}

func createHSMPartition(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant) (ctrl.Result, error) {
	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	hsmUrl := fmt.Sprintf("https://%s/apis/smd/hsm/v2/partitions", GetApiGateway())
	result, hsmPartitionBytes, err := buildHsmPartitionPayload(log, t, t.Spec.TenantResource.Xnames)
	if err != nil {
		return result, err
	}

	req, err := http.NewRequest(http.MethodPost, hsmUrl, bytes.NewBuffer(hsmPartitionBytes))
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
		log.Info("Created HSM partition: " + t.Spec.TenantResource.HsmPartitionName)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, errors.New("HSM returned a non-200 response creating partition")
}

func DeleteHSMGroup(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant) (ctrl.Result, error) {

	result, groupList, err := ListHSMGroups(ctx, log, t)
	if err != nil {
		return result, err
	}

	foundGroup := false
	for _, group := range groupList {
		if group.Label == t.Spec.TenantResource.HsmGroupLabel {
			foundGroup = true
			break
		}
	}

	if !foundGroup {
		log.Info("HSM group already deleted: " + t.Spec.TenantResource.HsmPartitionName)
		return ctrl.Result{}, nil
	}

	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	hsmUrl := fmt.Sprintf("https://%s/apis/smd/hsm/v2/groups/%s", GetApiGateway(), t.Spec.TenantResource.HsmGroupLabel)
	hsmGroup := HsmGroup{}
	hsmGroup.Label = t.Spec.TenantResource.HsmGroupLabel
	hsmGroupBytes, err := json.Marshal(hsmGroup)
	if err != nil {
		return ctrl.Result{}, err
	}

	req, err := http.NewRequest(http.MethodDelete, hsmUrl, bytes.NewBuffer(hsmGroupBytes))
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
		log.Info("Deleted HSM group: " + hsmGroup.Label)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, errors.New("HSM returned a non-200 response deleting group")
}

func DeleteHSMPartition(ctx context.Context, log logr.Logger, t *v1alpha1.Tenant) (ctrl.Result, error) {

	result, partitionList, err := ListHSMPartitions(ctx, log, t)
	if err != nil {
		return result, err
	}

	foundPartition := false
	for _, partition := range partitionList {
		if partition.Name == t.Spec.TenantResource.HsmPartitionName {
			foundPartition = true
			break
		}
	}

	if !foundPartition {
		log.Info("HSM partition already deleted: " + t.Spec.TenantResource.HsmPartitionName)
		return ctrl.Result{}, nil
	}

	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	hsmUrl := fmt.Sprintf("https://%s/apis/smd/hsm/v2/partitions/%s", GetApiGateway(), t.Spec.TenantResource.HsmPartitionName)
	hsmPartition := HsmPartition{}
	hsmPartition.Name = t.Spec.TenantResource.HsmPartitionName
	hsmPartitionBytes, err := json.Marshal(hsmPartition)
	if err != nil {
		return ctrl.Result{}, err
	}

	req, err := http.NewRequest(http.MethodDelete, hsmUrl, bytes.NewBuffer(hsmPartitionBytes))
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
		log.Info("Deleted HSM partition: " + hsmPartition.Name)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, errors.New("HSM returned a non-200 response deleting partition")
}
