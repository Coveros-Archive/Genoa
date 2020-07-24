package v3

import (
	"fmt"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"os"
	"time"
)

type UpgradeOptions struct {
	Namespace                string
	DryRun                   bool
	DisableHooks             bool
	Wait                     bool
	Timeout                  time.Duration
	ReleaseName              string
	Atomic                   bool
	SkipCRDs                 bool
	DisableOpenAPIValidation bool
	CleanupOnFail            bool
	Force                    bool
}

func (h *HelmV3) UpgradeRelease(chartPath string, opts UpgradeOptions, values map[string]interface{}) (*release.Release, error) {

	_, err := os.Stat(chartPath)
	if err != nil {
		return nil, err
	}

	loadedChart, errLoadingChart := loader.Load(chartPath)
	if errLoadingChart != nil {
		logger.Error(errLoadingChart, fmt.Sprintf("Failed to load %s chart", chartPath))
		return nil, errLoadingChart
	}

	upgradeAction := action.NewUpgrade(h.actionConfig)
	opts.setUpgradeOptions(upgradeAction)
	return upgradeAction.Run(opts.ReleaseName, loadedChart, values)

}

func (u UpgradeOptions) setUpgradeOptions(upgradeAction *action.Upgrade) {
	upgradeAction.Wait = u.Wait
	upgradeAction.DryRun = u.DryRun
	upgradeAction.DisableOpenAPIValidation = u.DisableOpenAPIValidation
	upgradeAction.SkipCRDs = u.SkipCRDs
	upgradeAction.Atomic = u.Atomic
	upgradeAction.Namespace = u.Namespace
	upgradeAction.DisableHooks = u.DisableHooks
	upgradeAction.Timeout = u.Timeout
	upgradeAction.CleanupOnFail = u.CleanupOnFail
	upgradeAction.Force = u.Force
}
