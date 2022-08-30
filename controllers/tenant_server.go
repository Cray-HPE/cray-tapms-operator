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

package controllers

import (
	"context"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/Cray-HPE/cray-tapms-operator/api/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TenantServer struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type ResponseError struct {
	Message string `json:"message"`
} //@name ResponseError

type ResponseOk struct {
	Message string `json:"message"`
} //@name ResponseOk

func (r *TenantServer) SetupServerController(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1.Tenant{}).
		Complete(r)
	if err != nil {
		return err
	}
	r.initRoutes()
	return nil
}

func (r *TenantServer) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *TenantServer) initRoutes() {
	router := gin.Default()
	router.GET("/apis/tapms/v1/tenants", r.GetTenants)
	router.GET("/apis/tapms/v1/tenant/:id", r.GetTenant)
	router.NoRoute(r.noRoute)
	go router.Run(v1.GetServerPort())
}

func (r *TenantServer) noRoute(c *gin.Context) {
	c.JSON(404, gin.H{"message": "Page not found"})
}

func (r *TenantServer) GetTenantsFromCache(c *gin.Context) (*v1.TenantList, error) {
	var tenantList v1.TenantList
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := r.List(ctx, &tenantList, &client.ListOptions{
		Namespace: "tenants",
	})

	if err != nil {
		if k8serrors.IsNotFound(err) {
			r.Log.Info("No tenants found")
		} else {
			return nil, fmt.Errorf("failed to get tenant list: %w", err)
		}
	}
	return &tenantList, nil
}

// GetTenants
// @Summary Get list of tenants
// @Tags    Tenant and Partition Management System
// @Accept  json
// @Produce json
// @Success 200 {object} ResponseOk
// @Failure 400 {object} ResponseError
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router  /apis/tapms/v1/tenants [get]
func (r *TenantServer) GetTenants(c *gin.Context) {
	tenantList, err := r.GetTenantsFromCache(c)
	if err != nil {
		c.JSON(500, ResponseError{Message: fmt.Sprint(err)})
		return
	}
	c.JSON(200, tenantList)
}

// GetTenant
// @Summary Get a tenant spec/status
// @Param   id path string true "either name or uuid of the tenant"
// @Tags    Tenant and Partition Management System
// @Accept  json
// @Produce json
// @Success 200 {object} ResponseOk
// @Failure 400 {object} ResponseError
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router  /apis/tapms/v1/tenant/{id} [get]
func (r *TenantServer) GetTenant(c *gin.Context) {
	id := c.Param("id")
	tenantList, err := r.GetTenantsFromCache(c)
	if err != nil {
		c.JSON(500, ResponseError{Message: fmt.Sprint(err)})
		return
	}

	if id == "" {
		c.JSON(400, ResponseError{Message: "Tenant name or UUID must be provided."})
		return
	}

	for _, tenant := range tenantList.Items {
		if tenant.Name == id || tenant.Status.UUID == id {
			c.JSON(200, tenant)
			return
		}
	}

	c.JSON(404, fmt.Sprintf("Tenant with name/uuid '%s' not found.", id))
}
