package manager

import (
	"flag"
	"time"

	"github.com/spf13/pflag"

	"github.com/mumoshu/okra/pkg/clclient"
	"github.com/mumoshu/okra/pkg/controllers"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

type Manager struct {
	MetricsAddr          string
	EnableLeaderElection bool
	SyncPeriod           time.Duration
}

func (m *Manager) AddFlags(fs flag.FlagSet) {
	fs.StringVar(&m.MetricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	fs.BoolVar(&m.EnableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	fs.DurationVar(&m.SyncPeriod, "sync-period", 30*time.Second, "Determines the minimum frequency at which K8s resources managed by this controller are reconciled.")

	//	flag.Parse()
}

func (m *Manager) AddPFlags(fs *pflag.FlagSet) {
	fs.StringVar(&m.MetricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	fs.BoolVar(&m.EnableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	fs.DurationVar(&m.SyncPeriod, "sync-period", 30*time.Second, "Determines the minimum frequency at which K8s resources managed by this controller are reconciled.")

	//	flag.Parse()
}

func (m *Manager) Run() error {
	var (
		err error
	)

	logger := zap.New(func(o *zap.Options) {
		o.Development = true
	})

	ctrl.SetLogger(logger)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             clclient.Scheme(),
		MetricsBindAddress: m.MetricsAddr,
		LeaderElection:     m.EnableLeaderElection,
		LeaderElectionID:   "okra",
		Port:               9443,
		SyncPeriod:         &m.SyncPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	clusterSetReconciler := &controllers.ClusterSetReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ClusterSet"),
		Scheme: mgr.GetScheme(),
	}

	if err = clusterSetReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterSet")
		return err
	}

	awsTargetGroupSetReconciler := &controllers.AWSTargetGroupSetReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AWSTargetGroupSet"),
		Scheme: mgr.GetScheme(),
	}

	if err = awsTargetGroupSetReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AWSTargetGroupSet")
		return err
	}

	awsALBConfigReconciler := &controllers.AWSApplicationLoadBalancerConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AWSApplicationLoadBalancerConfig"),
		Scheme: mgr.GetScheme(),
	}

	if err = awsALBConfigReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AWSApplicationLoadBalancerConfig")
		return err
	}

	cellReconciler := &controllers.CellReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Cell"),
		Scheme: mgr.GetScheme(),
	}

	if err = cellReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cell")
		return err
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}
