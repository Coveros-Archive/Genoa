package controllers

import (
	"context"
	"fmt"
	"github.com/coveros/genoa/api/v1alpha1"
	v3 "github.com/coveros/genoa/pkg/helm/v3"
	"github.com/coveros/genoa/pkg/utils"
	"helm.sh/helm/v3/pkg/release"
	"strings"
	"time"
)

func (r *ReleaseReconciler) cleanup(cr *v1alpha1.Release, actionConfig *v3.HelmV3) error {
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

	// fourth, check if other Releases exist in the same namespace
	hrList := &v1alpha1.ReleaseList{}
	if errGettingHrList := r.Client.List(context.TODO(), hrList); errGettingHrList != nil {
		return errGettingHrList
	}

	/**
	Finally, if we are allowed to delete the namespace AND there are NO OTHER Release's within, delete it.
	However this wont work if

	- A Release called "foo" gets installed in "foo" namespace with deleteNamespace annotation
	- A Release called "bar" gets installed in "bar" namespace without deleteNamespace annotation
	- Or if user has installed in their own custom existing namespace

	- if "foo" Release gets deleted, there will be 1 Release left in that namespace, but when
		"bar" release gets deleted it wont have the annotation to delete the namespace

	TODO: Maybe a better approach is to add a namespace label as soon as Release with deleteRelease annotation gets
	installed in a namespace -- this way we can keep a track of which namespace is supposed to be deleted such that
	when all Releases are gone, we can finally delete the namespace

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

func (r *ReleaseReconciler) pullChart(namespace, crName, repoAlias, chartName, version string, actionConfig *v3.HelmV3) (string, error) {

	// find repo url from repo config file
	repoUrl, username, password, errLookingUpRepo := actionConfig.GetRepoUrlFromRepoConfig(repoAlias)
	if errLookingUpRepo != nil {
		return "", errLookingUpRepo
	}

	// download chart
	chartPath, errDownloadingChart := actionConfig.DownloadChart(repoUrl, repoAlias,
		chartName, version,
		username, password,
		fmt.Sprintf("%v-%v", namespace, crName))
	if errDownloadingChart != nil {
		return "", errDownloadingChart
	}

	// return chart path
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

func getReleaseInstallOptions(cr *v1alpha1.Release) v3.InstallOptions {
	spec := cr.Spec
	installOptions := v3.InstallOptions{
		Namespace:                cr.GetNamespace(),
		DryRun:                   spec.DryRun,
		Wait:                     spec.Wait,
		Timeout:                  time.Duration(spec.WaitTimeout),
		ReleaseName:              cr.GetName(),
		DisableHooks:             spec.DisableHooks,
		DisableOpenAPIValidation: spec.DisableOpenAPIValidation,
		Atomic:                   spec.Atomic,
		IncludeCRDs:              spec.IncludeCRDs,
	}
	return installOptions
}

func getReleaseUpgradeOptions(cr *v1alpha1.Release) v3.UpgradeOptions {
	upgradeOpts := v3.UpgradeOptions{
		Namespace:                cr.GetNamespace(),
		DryRun:                   cr.Spec.DryRun,
		Wait:                     cr.Spec.Wait,
		Timeout:                  time.Duration(cr.Spec.WaitTimeout),
		ReleaseName:              cr.GetName(),
		DisableHooks:             cr.Spec.DisableHooks,
		DisableOpenAPIValidation: cr.Spec.DisableOpenAPIValidation,
		Atomic:                   cr.Spec.Atomic,
		CleanupOnFail:            cr.Spec.CleanupOnFail,
		SkipCRDs:                 !cr.Spec.IncludeCRDs,
		Force:                    cr.Spec.ForceUpgrade,
	}
	return upgradeOpts
}
