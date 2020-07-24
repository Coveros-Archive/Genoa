package v3

import (
	"errors"
	"fmt"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

func (h *HelmV3) UninstallRelease(releaseName string) (*release.UninstallReleaseResponse, error) {
	uninstallClient := action.NewUninstall(h.actionConfig)
	uninstallClient.DryRun = false
	uninstallClient.KeepHistory = false
	out, err := uninstallClient.Run(releaseName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			logger.Info(fmt.Sprintf("Looks like the %v release in namespace %v no longer exists... moving on", releaseName, h.namespace))
			return out, nil
		}
		return out, err
	}
	return out, nil
}
