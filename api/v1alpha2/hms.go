/*
 *
 *  MIT License
 *
 *  (C) Copyright 2022-2024 Hewlett Packard Enterprise Development LP
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
	Label          string
	Description    string
	ExclusiveGroup string
	Tags           []string
	Members        HsmIds
}
type HsmComponent struct {
	ID                  string
	Type                string
	State               string
	Flag                string
	Enabled             bool
	SoftwareStatus      string
	Role                string
	SubRole             string
	NID                 int32
	Subtype             string
	NetType             string
	Arch                string
	Class               string
	ReservationDisabled bool
	Locked              bool
}
type HsmComponentList struct {
	Components []HsmComponent
}

func ListHSMGroups(ctx context.Context, log logr.Logger) (ctrl.Result, []HsmGroup, error) {

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

func ListHSMPartitions(ctx context.Context, log logr.Logger) (ctrl.Result, []HsmPartition, error) {

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

func DetermineHSMGroupChanges(ctx context.Context, log logr.Logger, tenant *Tenant) (ctrl.Result, error) {
	//
	// First loop handles simple xname add/deletion from existing
	// resource (compute/application) or initial HSM group creation.
	//
	for _, resource := range tenant.Spec.TenantResources {
		if len(resource.HsmGroupLabel) > 0 {
			log.Info(fmt.Sprintf("Creating/updating HSM group for %s and resource type %s", tenant.Spec.TenantName, resource.Type))
			result, err := updateHSMGroup(ctx, log, tenant, resource)
			if err != nil {
				log.Error(err, "Failed to create/update HSM group")
				return result, err
			}
		}
	}

	//
	// Second loop handles case where a resource group is removed.
	//
	for _, statResource := range tenant.Status.TenantResources {
		haveSpec := false
		if len(statResource.HsmGroupLabel) > 0 {
			for _, resource := range tenant.Spec.TenantResources {
				if statResource.Type == resource.Type {
					haveSpec = true
				}
			}
		}
		if !haveSpec {
			result, err := editHsmGroupMembers(ctx, log, tenant.Name, statResource.HsmGroupLabel, statResource.Xnames, http.MethodDelete, statResource.EnforceExclusiveHsmGroups)
			if err != nil {
				log.Error(err, "Failed to delete HSM group members")
				return result, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func updateHSMGroup(ctx context.Context, log logr.Logger, t *Tenant, resource TenantResource) (ctrl.Result, error) {

	result, groupList, err := ListHSMGroups(ctx, log)
	if err != nil {
		return result, err
	}

	existingGroup := false
	for _, group := range groupList {
		if group.Label == resource.HsmGroupLabel {
			existingGroup = true
			break
		}
	}

	if !existingGroup {
		//
		// create the group
		//
		result, err := createHSMGroup(ctx, log, t.Name, resource.HsmGroupLabel, resource.Xnames, resource.EnforceExclusiveHsmGroups)
		if err != nil {
			return result, err
		}
		return ctrl.Result{}, nil
	} else {
		//
		// Check for any changes to update in the group
		//
		for _, statResource := range t.Status.TenantResources {
			if statResource.Type == resource.Type {
				log.Info(fmt.Sprintf("Checking for members deleted from HSM group %s and type %s", resource.HsmGroupLabel, resource.Type))
				deletedMembers := Difference(statResource.Xnames, resource.Xnames)
				result, err := editHsmGroupMembers(ctx, log, t.Name, resource.HsmGroupLabel, deletedMembers, http.MethodDelete, resource.EnforceExclusiveHsmGroups)
				if err != nil {
					log.Error(err, "Failed to delete HSM group members")
					return result, err
				}
				log.Info(fmt.Sprintf("Checking for members added from HSM group %s and type %s", resource.HsmGroupLabel, resource.Type))
				addedMembers := Difference(resource.Xnames, statResource.Xnames)
				result, err = editHsmGroupMembers(ctx, log, t.Name, resource.HsmGroupLabel, addedMembers, http.MethodPost, resource.EnforceExclusiveHsmGroups)
				if err != nil {
					log.Error(err, "Failed to add HSM group members")
					return result, err
				}
			}
		}

		//
		// Handles case where a resource group is added after the
		// HSM group is created.
		//
		for _, specResource := range t.Spec.TenantResources {
			if specResource.Type != resource.Type {
				continue
			}
			if len(specResource.HsmGroupLabel) > 0 {
				haveStatus := false
				for _, statResource := range t.Status.TenantResources {
					if specResource.Type == statResource.Type {
						haveStatus = true
					}
				}
				if !haveStatus {
					result, err := editHsmGroupMembers(ctx, log, t.Name, specResource.HsmGroupLabel, specResource.Xnames, http.MethodPost, specResource.EnforceExclusiveHsmGroups)
					if err != nil {
						log.Error(err, "Failed to add HSM group members")
						return result, err
					}
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func UpdateHSMPartition(ctx context.Context, log logr.Logger, t *Tenant, hsmPartitionName string, xnames []string) (ctrl.Result, error) {

	result, partitionList, err := ListHSMPartitions(ctx, log)
	if err != nil {
		return result, err
	}

	existingPartition := false
	for _, partition := range partitionList {
		if partition.Name == hsmPartitionName {
			existingPartition = true
			break
		}
	}

	if !existingPartition {
		//
		// create the partition
		//
		result, err := createHSMPartition(ctx, log, t.Name, hsmPartitionName, xnames)
		if err != nil {
			return result, err
		}
		return ctrl.Result{}, nil
	} else {
		//
		// Check for any changes to update in the partition
		//
		for _, specResource := range t.Spec.TenantResources {
			for _, statResource := range t.Status.TenantResources {
				if (statResource.Type == specResource.Type) && (specResource.HsmPartitionName == hsmPartitionName) {
					log.Info(fmt.Sprintf("Checking for members deleted from HSM partition %s", specResource.HsmPartitionName))
					deletedMembers := Difference(statResource.Xnames, specResource.Xnames)
					result, err := editHsmPartitionMembers(ctx, log, t.Name, hsmPartitionName, deletedMembers, http.MethodDelete)
					if err != nil {
						log.Error(err, "Failed to delete HSM partition members")
						return result, err
					}
				}
			}
		}
		for _, specResource := range t.Spec.TenantResources {
			for _, statResource := range t.Status.TenantResources {
				if (statResource.Type == specResource.Type) && (specResource.HsmPartitionName == hsmPartitionName) {
					log.Info(fmt.Sprintf("Checking for members added to HSM partition %s", specResource.HsmPartitionName))
					addedMembers := Difference(specResource.Xnames, statResource.Xnames)
					result, err = editHsmPartitionMembers(ctx, log, t.Name, hsmPartitionName, addedMembers, http.MethodPost)
					if err != nil {
						log.Error(err, "Failed to add HSM partition members")
						return result, err
					}
				}
			}
		}
	}
	return ctrl.Result{}, nil
}

func editHsmGroupMembers(ctx context.Context, log logr.Logger, tenantName string, hsmGroupLabel string, changedMembers []string, httpMethod string, enforceExclusiveHsmGroups bool) (ctrl.Result, error) {
	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	for _, member := range changedMembers {

		hsmUrl := ""
		action := ""
		hsmGroupBytes := []byte{}
		if httpMethod == http.MethodPost {
			hsmUrl = fmt.Sprintf("https://%s/apis/smd/hsm/v2/groups/%s/members", GetApiGateway(), hsmGroupLabel)
			action = "adding"
			hsmId := HsmMemberId{}
			hsmId.Id = member
			hsmGroupBytes, err = json.Marshal(hsmId)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if httpMethod == http.MethodDelete {
			hsmUrl = fmt.Sprintf("https://%s/apis/smd/hsm/v2/groups/%s/members/%s", GetApiGateway(), hsmGroupLabel, member)
			action = "removing"
			memberArray := []string{member}
			result, hsmGroupBytes, err = buildHsmGroupPayload(log, tenantName, hsmGroupLabel, memberArray, enforceExclusiveHsmGroups)
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
				log.Info(fmt.Sprintf("HSM member %s already deleted from group %s", member, hsmGroupLabel))
			} else if resp.StatusCode == 409 {
				log.Info(fmt.Sprintf("HSM member %s already added to group %s", member, hsmGroupLabel))
			} else {
				return ctrl.Result{}, fmt.Errorf("HSM returned a non-200 response %s member %s for group %s", action, member, hsmGroupLabel)
			}
		}
	}

	return ctrl.Result{}, nil
}

func editHsmPartitionMembers(ctx context.Context, log logr.Logger, tenantName string, hsmPartitionName string, changedMembers []string, httpMethod string) (ctrl.Result, error) {
	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	for _, member := range changedMembers {

		hsmUrl := ""
		action := ""
		hsmPartitionBytes := []byte{}
		if httpMethod == http.MethodPost {
			hsmUrl = fmt.Sprintf("https://%s/apis/smd/hsm/v2/partitions/%s/members", GetApiGateway(), hsmPartitionName)
			action = "adding"
			hsmId := HsmMemberId{}
			hsmId.Id = member
			hsmPartitionBytes, err = json.Marshal(hsmId)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if httpMethod == http.MethodDelete {
			hsmUrl = fmt.Sprintf("https://%s/apis/smd/hsm/v2/partitions/%s/members/%s", GetApiGateway(), hsmPartitionName, member)
			action = "removing"
			memberArray := []string{member}
			result, hsmPartitionBytes, err = buildHsmPartitionPayload(log, tenantName, hsmPartitionName, memberArray)
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
				log.Info(fmt.Sprintf("HSM member %s already deleted from partition %s", member, hsmPartitionName))
			} else if resp.StatusCode == 409 {
				log.Info(fmt.Sprintf("HSM member %s already added to partition %s", member, hsmPartitionName))
			} else {
				return ctrl.Result{}, fmt.Errorf("HSM returned a non-200 response %s member %s for partition %s", action, member, hsmPartitionName)
			}
		}
	}

	return ctrl.Result{}, nil
}

func buildHsmPartitionPayload(log logr.Logger, tenantName string, hsmPartitionName string, xnames []string) (ctrl.Result, []byte, error) {

	hsmPartition := HsmPartition{}
	hsmPartition.Name = hsmPartitionName
	hsmPartition.Tags = append(hsmPartition.Tags, tenantName)
	hsmPartition.Members.Ids = append(hsmPartition.Members.Ids, xnames...)
	hsmPartitionBytes, err := json.Marshal(hsmPartition)
	if err != nil {
		return ctrl.Result{}, nil, err
	}
	return ctrl.Result{}, hsmPartitionBytes, err
}

func buildHsmGroupPayload(log logr.Logger, tenantName string, hsmGroupLabel string, xnames []string, enforceExclusiveHsmGroups bool) (ctrl.Result, []byte, error) {

	hsmGroup := HsmGroup{}
	hsmGroup.Label = hsmGroupLabel
	if enforceExclusiveHsmGroups {
		hsmGroup.ExclusiveGroup = "tapms-exclusive-group-label"
	} else {
		hsmGroup.ExclusiveGroup = ""
	}
	hsmGroup.Tags = append(hsmGroup.Tags, tenantName)
	for _, xname := range xnames {
		hsmGroup.Members.Ids = append(hsmGroup.Members.Ids, xname)
	}
	hsmGroupBytes, err := json.Marshal(hsmGroup)
	if err != nil {
		return ctrl.Result{}, nil, err
	}
	return ctrl.Result{}, hsmGroupBytes, err
}

func createHSMGroup(ctx context.Context, log logr.Logger, tenantName string, hsmGroupLabel string, xnames []string, enforceExclusiveHsmGroups bool) (ctrl.Result, error) {
	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	hsmUrl := fmt.Sprintf("https://%s/apis/smd/hsm/v2/groups", GetApiGateway())
	result, hsmGroupBytes, err := buildHsmGroupPayload(log, tenantName, hsmGroupLabel, xnames, enforceExclusiveHsmGroups)
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
		log.Info("Created HSM group: " + hsmGroupLabel)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, errors.New("HSM returned a non-200 response creating group")
}

func createHSMPartition(ctx context.Context, log logr.Logger, tenantName string, hsmPartitionName string, xnames []string) (ctrl.Result, error) {
	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	hsmUrl := fmt.Sprintf("https://%s/apis/smd/hsm/v2/partitions", GetApiGateway())
	result, hsmPartitionBytes, err := buildHsmPartitionPayload(log, tenantName, hsmPartitionName, xnames)
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
		log.Info("Created HSM partition: " + hsmPartitionName)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, errors.New("HSM returned a non-200 response creating partition")
}

func DeleteHSMGroup(ctx context.Context, log logr.Logger, hsmGroupLabel string) (ctrl.Result, error) {

	result, groupList, err := ListHSMGroups(ctx, log)
	if err != nil {
		return result, err
	}

	foundGroup := false
	for _, group := range groupList {
		if group.Label == hsmGroupLabel {
			foundGroup = true
			break
		}
	}

	if !foundGroup {
		log.Info("HSM group already deleted: " + hsmGroupLabel)
		return ctrl.Result{}, nil
	}

	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	hsmUrl := fmt.Sprintf("https://%s/apis/smd/hsm/v2/groups/%s", GetApiGateway(), hsmGroupLabel)
	hsmGroup := HsmGroup{}
	hsmGroup.Label = hsmGroupLabel
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

func DeleteHSMPartition(ctx context.Context, log logr.Logger, hsmPartitionName string) (ctrl.Result, error) {

	result, partitionList, err := ListHSMPartitions(ctx, log)
	if err != nil {
		return result, err
	}

	foundPartition := false
	for _, partition := range partitionList {
		if partition.Name == hsmPartitionName {
			foundPartition = true
			break
		}
	}

	if !foundPartition {
		log.Info("HSM partition already deleted: " + hsmPartitionName)
		return ctrl.Result{}, nil
	}

	result, token, err := GetToken(ctx, log, false)
	if err != nil {
		return result, err
	}

	hsmUrl := fmt.Sprintf("https://%s/apis/smd/hsm/v2/partitions/%s", GetApiGateway(), hsmPartitionName)
	hsmPartition := HsmPartition{}
	hsmPartition.Name = hsmPartitionName
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

func GetComponentList(ctx context.Context, log logr.Logger, nodeType string, role string) (*HsmComponentList, error) {

	_, token, err := GetToken(ctx, log, false)
	if err != nil {
		return nil, err
	}
	hsmUrl := fmt.Sprintf("https://%s/apis/smd/hsm/v2/State/Components?type=%s&role=%s", GetApiGateway(), nodeType, role)
	hsmComponentList := HsmComponentList{}
	hsmComponentListBytes, err := json.Marshal(hsmComponentList)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, hsmUrl, bytes.NewBuffer(hsmComponentListBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	HTTPClient := NewHttpClient()
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.New("HSM returned a non-200 response listing components")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &hsmComponentList)
	if err != nil {
		return nil, err
	}

	return &hsmComponentList, nil
}
