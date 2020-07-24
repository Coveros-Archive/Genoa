package v3

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"os"
	"time"
)

type InstallOptions struct {
	Namespace                string
	DryRun                   bool
	DisableHooks             bool
	Wait                     bool
	Timeout                  time.Duration
	ReleaseName              string
	Atomic                   bool
	SkipCRDs                 bool
	DisableOpenAPIValidation bool
	IncludeCRDs              bool
}

//InstallRelease installs helm charts, assuming a chart path locally exists
func (h *HelmV3) InstallRelease(chartPath string, options InstallOptions, values map[string]interface{}) (*release.Release, error) {

	_, err := os.Stat(chartPath)
	if err != nil {
		return nil, err
	}

	logger.Info("loading chart")
	loadedChart, errLoadingChart := loader.Load(chartPath)
	if errLoadingChart != nil {
		return nil, errLoadingChart
	}

	logger.Info("installing chart")
	installAction := action.NewInstall(h.actionConfig)
	options.setInstallOptions(installAction)
	return installAction.Run(loadedChart, values)
}

func (i InstallOptions) setInstallOptions(installAction *action.Install) *action.Install {
	installAction.DryRun = i.DryRun
	installAction.DisableHooks = i.DisableHooks
	installAction.Wait = i.Wait
	installAction.Timeout = i.Timeout * time.Second
	installAction.Namespace = i.Namespace
	installAction.ReleaseName = i.ReleaseName
	installAction.Atomic = i.Atomic
	installAction.SkipCRDs = i.SkipCRDs
	installAction.DisableOpenAPIValidation = i.DisableOpenAPIValidation
	installAction.IncludeCRDs = i.IncludeCRDs
	return installAction
}
