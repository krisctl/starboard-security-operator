package main

import (
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/aquasecurity/starboard-security-operator/pkg/reports"

	"github.com/aquasecurity/starboard-security-operator/pkg/aqua/scanner"

	"github.com/aquasecurity/starboard-security-operator/pkg/controllers"
	"github.com/aquasecurity/starboard-security-operator/pkg/etc"
	"github.com/aquasecurity/starboard/pkg/find/vulnerabilities"
	"github.com/aquasecurity/starboard/pkg/find/vulnerabilities/trivy"
	"github.com/aquasecurity/starboard/pkg/kube"
	pods "github.com/aquasecurity/starboard/pkg/kube/pod"
	"k8s.io/client-go/kubernetes"

	starboardv1alpha1 "github.com/aquasecurity/starboard/pkg/apis/aquasecurity/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	// GoReleaser sets three ldflags:
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	versionInfo = etc.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = corev1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = starboardv1alpha1.AddToScheme(scheme)
}

func main() {
	logf.SetLogger(zap.New())

	ctrl.SetLogger(logf.Log.WithName("starboard-operator"))
	if err := run(); err != nil {
		setupLog.Error(err, "Unable to run manager")
	}
}

func run() error {
	config, err := etc.GetConfig()
	if err != nil {
		return err
	}

	kubernetesConfig := ctrl.GetConfigOrDie()
	// TODO Do not use this client unless absolutely necessary. We should rely on the client constructed by the ctrl.NewManager()
	kubernetesClientset, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return err
	}
	pods := pods.NewPodManager(kubernetesClientset)

	scanner, err := getEnabledScanner(config, kubernetesClientset, pods)
	if err != nil {
		return err
	}

	mgr, err := ctrl.NewManager(kubernetesConfig, ctrl.Options{
		Scheme: scheme,
	})

	if err != nil {
		return fmt.Errorf("unable to start manager: %w", err)
	}

	store := reports.NewStore(mgr.GetClient(), scheme)

	if err = (&controllers.PodReconciler{
		StarboardNamespace: config.Operator.StarboardNamespace,
		Namespace:          config.Operator.Namespace,
		Client:             mgr.GetClient(),
		Store:              store,
		Scanner:            scanner,
		Log:                ctrl.Log.WithName("controllers").WithName("pod"),
		Scheme:             mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create pod controller: %w", err)
	}

	if err = (&controllers.JobReconciler{
		StarboardNamespace: config.Operator.StarboardNamespace,
		Client:             mgr.GetClient(),
		Store:              store,
		Scanner:            scanner,
		Pods:               pods,
		Log:                ctrl.Log.WithName("controllers").WithName("job"),
		Scheme:             mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create job controller: %w", err)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}

func getEnabledScanner(config etc.Config, kubeClientset kubernetes.Interface, pods *pods.Manager) (vulnerabilities.ScannerAsync, error) {
	if config.ScannerTrivy.Enabled && config.ScannerAquaCSP.Enabled {
		return nil, fmt.Errorf("invalid configuration: multiple vulnerability scanners enabled")
	}
	if !config.ScannerTrivy.Enabled && !config.ScannerAquaCSP.Enabled {
		return nil, fmt.Errorf("invalid configuration: none vulnerability scanner enabled")
	}
	if config.ScannerTrivy.Enabled {
		setupLog.Info("Using Trivy as vulnerability scanner", "version", config.ScannerTrivy.Version)
		return trivy.NewScanner(kube.ScannerOpts{}, kubeClientset), nil
	}
	if config.ScannerAquaCSP.Enabled {
		setupLog.Info("Using Aqua CSP as vulnerability scanner", "version", config.ScannerAquaCSP.Version)
		return scanner.NewScanner(versionInfo, config, &scanner.RandomNamesGenerator{}, pods), nil
	}
	return nil, errors.New("invalid configuration: unhandled vulnerability scanners config")
}
