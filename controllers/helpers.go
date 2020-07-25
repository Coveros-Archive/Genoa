package controllers

import (
	"context"
	"coveros.com/api/v1alpha1"
	v3 "coveros.com/pkg/helm/v3"
	"coveros.com/pkg/utils"
	"fmt"
	"helm.sh/helm/v3/pkg/release"
	"strings"
)

func (r *HelmReleaseReconciler) cleanup(cr *v1alpha1.HelmRelease, actionConfig *v3.HelmV3) error {
	var deleteNamespace bool

	// first, delete the helm release
	if _, errUninstallingRelease := actionConfig.UninstallRelease(cr.GetName()); errUninstallingRelease != nil {
		return errUninstallingRelease
	}

	// second, check if we can delete the namespace
	if val, ok := cr.GetAnnotations()[utils.AutoDeleteNamespaceAnnotation]; ok && strings.ToLower(val) == "true" {
		deleteNamespace = true
	}

	// third, remove finalizer from CR
	if errRemovingFinalizer := utils.RemoveFinalizer(utils.ReleaseFinalizer, r.Client, cr); errRemovingFinalizer != nil {
		return errRemovingFinalizer
	}

	// fourth, check if other helmReleases exist in the same namespace
	hrList := &v1alpha1.HelmReleaseList{}
	if errGettingHrList := r.Client.List(context.TODO(), hrList); errGettingHrList != nil {
		return errGettingHrList
	}

	/**
	Finally, if we are allowed to delete the namespace AND there are NO OTHER HelmRelease's within, delete it.
	However this wont work if

	- A helmRelease called "foo" gets installed in "foo" namespace with deleteNamespace annotation
	- A helmRelease called "bar" gets installed in "bar" namespace without deleteNamespace annotation
	- Or if user has installed in their own custom existing namespace

	- if "foo" helmRelease gets deleted, there will be 1 helmRelease left in that namespace, but when
		"bar" release gets deleted it wont have the annotation to delete the namespace

	TODO: Maybe a better approach is to add a namespace label as soon as helmRelease with deleteRelease annotation gets
	installed in a namespace -- this way we can keep a track of which namespace is supposed to be deleted such that
	when all helmReleases are gone, we can finally delete the namespace

	*/
	if deleteNamespace && len(hrList.Items) == 0 {
		//if errDeletingNamespace := r.Client.Delete(context.TODO(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: cr.GetNamespace()}}); errDeletingNamespace != nil {
		//	if errors.IsNotFound(errDeletingNamespace) {
		//		return nil
		//	}
		//	return errDeletingNamespace
		//}
	}
	return nil
}

func (r *HelmReleaseReconciler) pullChart(namespace, crName, repoAlias, chartName, version string, actionConfig *v3.HelmV3) (string, error) {
	repoUrl, username, password, errLookingUpRepo := actionConfig.GetRepoUrlFromRepoConfig(repoAlias)
	if errLookingUpRepo != nil {
		return "", errLookingUpRepo
	}
	r.Log.Info(fmt.Sprintf("downloading chart from %s", repoUrl))
	chartPath, errDownloadingChart := actionConfig.DownloadChart(repoUrl,
		chartName, version,
		username, password,
		fmt.Sprintf("%v-%v", namespace, crName))
	if errDownloadingChart != nil {
		return "", errDownloadingChart
	}
	return chartPath, nil
}

func isReleasePending(releaseInfo *release.Release) bool {
	if releaseInfo.Info.Status == release.StatusPendingInstall ||
		releaseInfo.Info.Status == release.StatusUninstalling ||
		releaseInfo.Info.Status == release.StatusPendingRollback ||
		releaseInfo.Info.Status == release.StatusPendingUpgrade {
		return true
	}
	return false
}
