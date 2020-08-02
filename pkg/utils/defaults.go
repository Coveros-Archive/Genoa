package utils

const (
	ReleaseFinalizer              = "coveros.apps.Release"
	AutoDeleteNamespaceAnnotation = ReleaseFinalizer + "/autoDeleteNamespace"
	GitBranchToFollowAnnotation   = ReleaseFinalizer + "/follow-git-branch"
)
