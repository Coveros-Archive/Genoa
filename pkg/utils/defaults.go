package utils

const (
	ReleaseFinalizer                = "coveros.apps.genoa"
	AutoDeleteNamespaceAnnotation   = ReleaseFinalizer + "/autoDeleteNamespace"
	GitBranchToFollowAnnotation     = ReleaseFinalizer + "/follow-git-branch"
	SlackChannelIDAnnotation        = ReleaseFinalizer + "/notification-channel-id"
	EnvVarNotificationProvider      = "NOTIFICATION_PROVIDER"
	EnvVarNotificationProviderToken = "NOTIFICATION_PROVIDER_TOKEN"
	EnvVarGitlabSelfHostedUrl       = "GITLAB_SELF_HOSTED_URL"
	EnvVarGithubEnterpriseHostedUrl = "GITHUB_ENTERPRISE_HOSTED_URL"
)
