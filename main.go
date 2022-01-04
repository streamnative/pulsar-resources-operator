// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

package main

import (
	"flag"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	k8szap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	//+kubebuilder:scaffold:imports
	"github.com/streamnative/pulsar-operators/commons/pkg/trace"
	pulsarv1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/controllers"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(pulsarv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := k8szap.Options{
		Development: true,
	}
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

	ctrl.SetLogger(k8szap.New(k8szap.UseFlagOptions(&opts), k8szap.Encoder(encoder)))

	if err := trace.InitTracingProvider(); err != nil {
		setupLog.Error(err, "unable to set cloud tracing provider")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		// TODO uncomment it until kube-instrumentation upgrade controller-runtime version to newer
		// NewClient:              otelcontroller.NewClient,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "feaa54b6.streamnative.io",
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
