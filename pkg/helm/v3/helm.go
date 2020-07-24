package v3

import (
	"fmt"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/util"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type HelmV3 struct {
	namespace    string
	settings     *cli.EnvSettings
	actionConfig *action.Configuration
}

var logger = logf.Log.WithName("helmV3")

func DefaultEnvSettings() *cli.EnvSettings {
	return cli.New()
}

// this hack so we can continue using controller-runtime logger for consistency in logs..
func helmInfoLogF(format string, a ...interface{}) {
	logger.Info(fmt.Sprintf(format, a...))
}

func NewActionConfig(namespace string, cfg *rest.Config) (*HelmV3, error) {
	// the reason why I could not just use helm cli package to create action config...
	//actionConfig := &action.Configuration{}
	//if err := actionConfig.Init(DefaultEnvSettings().RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), logger.Infof); err != nil {
	//	return nil, err
	//}
	// https://github.com/helm/helm/issues/7845
	cfgFlagsForRestClient := &genericclioptions.ConfigFlags{
		Namespace:   &namespace,
		APIServer:   &cfg.Host,
		CAFile:      &cfg.CAFile,
		BearerToken: &cfg.BearerToken,
	}

	client := &kube.Client{
		Factory: util.NewFactory(cfgFlagsForRestClient),
		Log:     helmInfoLogF,
	}
	// clientSet is needed to initialize a release storage driver, or else we get nil pointer deference when listing or getting releases
	clientSet, err := client.Factory.KubernetesClientSet()
	if err != nil {
		return nil, err
	}

	// TODO: make this configurable
	secretsHelmDriver := driver.NewSecrets(clientSet.CoreV1().Secrets(namespace))
	secretsHelmDriver.Log = helmInfoLogF
	secretsHelmStorage := storage.Init(secretsHelmDriver)

	actionConfig := &action.Configuration{
		RESTClientGetter: cfgFlagsForRestClient,
		KubeClient:       client,
		Releases:         secretsHelmStorage,
		Log:              helmInfoLogF,
	}

	return &HelmV3{
		namespace:    namespace,
		settings:     DefaultEnvSettings(),
		actionConfig: actionConfig,
	}, nil
}
