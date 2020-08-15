package v3

import (
	"helm.sh/helm/v3/pkg/action"
	"time"
)

type RollbackToRevisionOptions struct {
	Force       bool
	Wait        bool
	WaitTimeout int
	ToRevision  int
}

func (h *HelmV3) RollbackToRevision(releaseName string, rollbackOptions RollbackToRevisionOptions) error {
	rollBackClient := action.NewRollback(h.actionConfig)
	rollbackOptions.withRollbackOptions(rollBackClient)
	return rollBackClient.Run(releaseName)
}

func (r RollbackToRevisionOptions) withRollbackOptions(rollbackClient *action.Rollback) {
	rollbackClient.Version = r.ToRevision
	rollbackClient.Wait = r.Wait
	rollbackClient.Timeout = time.Duration(r.WaitTimeout) * time.Second
	rollbackClient.Force = r.Force
}
