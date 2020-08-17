package utils

const (
	ReleaseFinalizer              = "coveros.apps.genoa"
	AutoDeleteNamespaceAnnotation = ReleaseFinalizer + "/autoDeleteNamespace"
	GitBranchToFollowAnnotation   = ReleaseFinalizer + "/follow-git-branch"
	SlackChannelIDAnnotation      = ReleaseFinalizer + "/slack-channel-id"
)
