package v3

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
)

func (h *HelmV3) GetRelease(releaseName string) (*release.Release, error) {
	getAction := action.NewGet(h.actionConfig)
	return getAction.Run(releaseName)
}
