package utils

const (
	ReleaseFinalizer                = "coveros.apps.genoa"
	AutoDeleteNamespaceAnnotation   = ReleaseFinalizer + "/autoDeleteNamespace"
	GitBranchToFollowAnnotation     = ReleaseFinalizer + "/follow-git-branch"
	SlackChannelIDAnnotation        = ReleaseFinalizer + "/notification-channel-id"
	EnvVarNotificationProvider      = "NOTIFICATION_PROVIDER"
	EnvVarNotificationProviderToken = "NOTIFICATION_PROVIDER_TOKEN"
)
