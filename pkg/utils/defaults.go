package utils

const (
	ReleaseFinalizer              = "coveros.apps.helmRelease"
	AutoDeleteNamespaceAnnotation = ReleaseFinalizer + "/autoDeleteNamespace"
	GitBranchToFollowAnnotation   = ReleaseFinalizer + "/follow-git-branch"
)
