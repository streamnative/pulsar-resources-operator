// Copyright 2024 StreamNative
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main is the entrypoint
package main

import (
	"flag"
	"os"
	"time"

	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	"github.com/streamnative/pulsar-resources-operator/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.) to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	k8szap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	//+kubebuilder:scaffold:imports
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/controllers"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(resourcev1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var resyncPeriod int
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&resyncPeriod, "resync-period", 10, "resyncPeriod is the base frequency the informers are resynced.")
	opts := k8szap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	var encoder zapcore.Encoder
	// If development is false, then will use json encoder and info log level
	// otherwise, will use console encoder and debug log level
	if !opts.Development {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.LevelKey = "severity"
		encoderConfig.MessageKey = "message"
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	if err := feature.SetFeatureGates(); err != nil {
		setupLog.Error(err, "failed to set feature gates")
		os.Exit(1)
	}

	ctrl.SetLogger(k8szap.New(k8szap.UseFlagOptions(&opts), k8szap.Encoder(encoder)))

	reSync := time.Duration(resyncPeriod) * time.Hour

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		// TODO uncomment it until kube-instrumentation upgrade controller-runtime version to newer
		// NewClient:              otelcontroller.NewClient,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "feaa54b6.streamnative.io",
		SyncPeriod:             &reSync,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// TODO get MaxConcurrentReconciles from cmd params
	if err = (&controllers.PulsarConnectionReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		Log:                ctrl.Log.WithName("controllers").WithName("PulsarConnection"),
		Recorder:           mgr.GetEventRecorderFor("pulsarconnection-controller"),
		PulsarAdminCreator: admin.NewPulsarAdmin,
		Retryer:            utils.NewReconcileRetryer(5, utils.NewEventSource(ctrl.Log.WithName("eventSource"))),
	}).SetupWithManager(mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PulsarConnection")
		os.Exit(1)
	}
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
