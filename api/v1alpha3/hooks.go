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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-logr/logr"
)

var validEventTypes = []string{"CREATE", "UPDATE", "DELETE"}

type TenantEventPayload struct {
	TenantSpec TenantSpec `json:"tenantspec"`
	EventType  string     `json:"eventtype"`
}

func CallHooks(tenant *Tenant, log logr.Logger, event string) error {

	for _, hook := range tenant.Spec.TenantHooks {
		err := validateEventType(log, hook.EventTypes)
		if err != nil {
			return err
		}

		if !Contains(hook.EventTypes, event) {
			continue
		}

		payload := TenantEventPayload{}
		payload.EventType = event
		payload.TenantSpec = tenant.Spec
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodPost, hook.Url, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return err
		}
		req.Header.Add("Content-Length", strconv.FormatInt(req.ContentLength, 10))

		blockText := ""
		if hook.BlockingCall {
			blockText = "Blocking"
		} else {
			blockText = "Notify"
		}

		Log.Info(fmt.Sprintf("Calling hook named '%s' at url %s (%s)", hook.Name, hook.Url, blockText))

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Block", strconv.FormatBool(hook.BlockingCall))

		HTTPClient := NewHttpClient()
		resp, err := HTTPClient.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return fmt.Errorf("%s call to '%s' hook at url %s returned a non-200 response code: %d", blockText, hook.Name, hook.Url, resp.StatusCode)
		}
		Log.Info(fmt.Sprintf("%s call to '%s' hook at url %s called successfully", blockText, hook.Name, hook.Url))
	}

	return nil
}

func validateEventType(log logr.Logger, events []string) error {
	for _, event := range events {
		if !Contains(validEventTypes, event) {
			return fmt.Errorf("EventType '%s' is not supported, must be one of: %v", event, validEventTypes)
		}
	}
	return nil
}
