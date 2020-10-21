package pkg

const (
	ReleaseFinalizer                = "coveros.apps.genoa"
	AutoDeleteNamespaceAnnotation   = ReleaseFinalizer + "/autoDeleteNamespace"
	GitBranchToFollowAnnotation     = ReleaseFinalizer + "/follow-git-branch"
	SlackChannelIDAnnotation        = ReleaseFinalizer + "/notification-channel-id"
	GitProvider                     = ReleaseFinalizer + "/git-provider"
	ReleaseFilePath                 = ReleaseFinalizer + "/git-release-file"
	GitOwnerRepo                    = ReleaseFinalizer + "/git-owner-repo"
	SkipChartSync                   = ReleaseFinalizer + "/skip-chart-sync"
	EnvVarNotificationProvider      = "NOTIFICATION_PROVIDER"
	EnvVarNotificationProviderToken = "NOTIFICATION_PROVIDER_TOKEN"
	EnvVarGitlabSelfHostedUrl       = "GITLAB_SELF_HOSTED_URL"
	EnvVarGithubEnterpriseHostedUrl = "GITHUB_ENTERPRISE_HOSTED_URL"
	GithubEventHeaderKey            = "X-Github-Event"
	GitlabEventHeaderKey            = "X-Gitlab-Event"
	EnvVarGithubReleaseFilesDir     = "GITHUB_DEPLOY_DIRECTORY"
	EnvVarGitlabReleaseFilesDir     = "GITLAB_DEPLOY_DIRECTORY"
	EnvVarGithubWebhookSecret       = "GITHUB_WEBHOOK_SECRET"
	EnvVarGitlabWebhookSecret       = "GITLAB_WEBHOOK_SECRET"
	EnvVarGithubPersonalAccessToken = "GITHUB_PERSONAL_ACCESS_TOKEN"
	EnvVarGitlabPersonalAccessToken = "GITLAB_PERSONAL_ACCESS_TOKEN"
)
