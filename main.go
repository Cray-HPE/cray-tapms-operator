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

//	@title			TAPMS Tenant Status API
//	@version		v1alpha3
//	@description	Read-Only APIs to Retrieve Tenant Status
//	@host			cray-tapms
//	@BasePath		/apis/tapms/

package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	hncapi "sigs.k8s.io/hierarchical-namespaces/api/v1alpha2"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v1alpha2 "github.com/Cray-HPE/cray-tapms-operator/api/v1alpha2"
	v1alpha3 "github.com/Cray-HPE/cray-tapms-operator/api/v1alpha3"
	"github.com/Cray-HPE/cray-tapms-operator/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha2.AddToScheme(scheme))
	utilruntime.Must(v1alpha3.AddToScheme(scheme))
	utilruntime.Must(hncapi.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var namespace string
	var webhookServerPort int
	flag.IntVar(&webhookServerPort, "webhook-server-port", 443, "The port that the webhook server serves at.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&namespace, "namespace", "tenants", "The namespace to watch for CRs in")
	flag.BoolVar(&enableLeaderElection, "leader-elect", true,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Namespace:              namespace,
		MetricsBindAddress:     metricsAddr,
		Port:                   webhookServerPort,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "7ff2bf92.hpe.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.TenantReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Tenants"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Tenants")
		os.Exit(1)
	}

	if err = (&controllers.TenantServer{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Server"),
		Scheme: mgr.GetScheme(),
	}).SetupServerController(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "server", "Server")
		os.Exit(1)
	}

	if err = (&controllers.TenantReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Tenants"),
		Scheme: mgr.GetScheme(),
	}).BuildRootTreeStructure(mgr); err != nil {
		setupLog.Error(err, "unable build initial tree structure")
		os.Exit(1)
	}

	if err = (&v1alpha3.Tenant{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Tenant")
		os.Exit(1)
	}

	v1alpha3.HooksClient = mgr.GetClient()

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
