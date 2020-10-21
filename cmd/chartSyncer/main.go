package main

import (
	"coveros.com/api/v1alpha1"
	"coveros.com/pkg/gitSync"
	v3 "coveros.com/pkg/helm/v3"
	"flag"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme = runtime.NewScheme()
	logger = ctrl.Log.WithName("chartSyncer")
)

func init() {
	clientgoscheme.AddToScheme(scheme)
	v1alpha1.AddToScheme(scheme)
}

func main() {
	var customRepoConfigPath string
	flag.StringVar(&customRepoConfigPath, "custom-helm-repos-file", "", "Your own custom helm repo files")

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	// create a controller-runtime client
	// and add my custom types to the scheme along with k8s client-go scheme ( k8s native types )
	cfg := config.GetConfigOrDie()
	client, err := ctrlClient.New(cfg, ctrlClient.Options{Scheme: scheme})
	if err != nil {
		logger.Error(err, "Failed to setup a controller-runtime client")
		os.Exit(1)
	}

	defaultHelmV3, err := v3.NewActionConfig("default", cfg)
	if err != nil {
		logger.Error(err, "Failed to create helm v3 client")
		os.Exit(1)
	}

	if customRepoConfigPath != "" {
		if errAddingCustomRepos := v3.AddReposFromFile(customRepoConfigPath); errAddingCustomRepos != nil {
			logger.Error(errAddingCustomRepos, "Failed to add custom helm repos")
			os.Exit(1)
		}
	}

	chartVersionUpdater := gitSync.ChartVersionSync{
		Client: client,
		Logger: logger,
		HelmV3: defaultHelmV3,
	}

	chartVersionUpdater.StartSync()

}
