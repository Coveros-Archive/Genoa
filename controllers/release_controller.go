/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"coveros.com/pkg"
	v3 "coveros.com/pkg/helm/v3"
	"coveros.com/pkg/utils"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/storage/driver"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"os"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"strings"
	"time"

	coverosv1alpha1 "coveros.com/api/v1alpha1"
)

// ReleaseReconciler reconciles a Release object
type ReleaseReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Cfg    *rest.Config
}

func (r *ReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: 7}).
		For(&coverosv1alpha1.Release{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=coveros.apps.com,resources=Releases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=coveros.apps.com,resources=Releases/status,verbs=get;update;patch
func (r *ReleaseReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("Release", req.NamespacedName)

	// Fetch the Release CR
	cr := &coverosv1alpha1.Release{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, cr)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			// do not requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}
	hrName := cr.GetName()
	hrNamespace := cr.GetNamespace()
	repoWithChartName := strings.SplitN(cr.Spec.Chart, "/", 2)
	var justChartName = repoWithChartName[1]
	if strings.Contains(justChartName, "/") {
		justChartName = strings.Split(justChartName, "/")[1]
	}
	repoAlias, chartName := repoWithChartName[0], repoWithChartName[1]
	helmV3, errCreatingActionConfig := v3.NewActionConfig(cr.GetNamespace(), r.Cfg)
	if errCreatingActionConfig != nil {
		return ctrl.Result{}, errCreatingActionConfig
	}

	// add finalizer
	if errAddingFinalizer := utils.AddFinalizer(utils.ReleaseFinalizer, r.Client, cr); errAddingFinalizer != nil {
		return ctrl.Result{}, errAddingFinalizer
	}

	// handle delete
	if cr.GetDeletionTimestamp() != nil {
		if errCleaningUp := r.cleanup(cr, helmV3); errCleaningUp != nil {
			return ctrl.Result{}, errCleaningUp
		}
		return ctrl.Result{}, nil // do not requeue
	}

	releaseInfo, errGettingReleaseInfo := helmV3.GetRelease(hrName)
	if errGettingReleaseInfo != nil {
		if errors.Is(errGettingReleaseInfo, driver.ErrReleaseNotFound) {
			r.Log.Info("release not found, installing now...")

			chartPath, errPullingChart := r.pullChart(hrNamespace, hrName, repoAlias, chartName, cr.Spec.Version, helmV3)
			if errPullingChart != nil {
				if _, ok := errPullingChart.(pkg.ErrorHelmRepoNeedsRefresh); ok {
					return ctrl.Result{Requeue: true}, helmV3.RefreshRepoIndex(repoAlias)
				}
				return ctrl.Result{}, errPullingChart
			}

			defer os.RemoveAll(strings.Split(chartPath, "/")[0])

			installOptions := v3.InstallOptions{
				Namespace:                hrNamespace,
				DryRun:                   cr.Spec.DryRun,
				Wait:                     cr.Spec.Wait,
				Timeout:                  time.Duration(cr.Spec.WaitTimeout),
				ReleaseName:              hrName,
				DisableHooks:             cr.Spec.DisableHooks,
				DisableOpenAPIValidation: cr.Spec.DisableOpenAPIValidation,
				Atomic:                   cr.Spec.Atomic,
				IncludeCRDs:              cr.Spec.IncludeCRDs,
			}
			r.Log.Info(fmt.Sprintf("%v: downloaded chart at %v", req.NamespacedName, chartPath))
			_, errInstallingChart := helmV3.InstallRelease(chartPath, installOptions, cr.Spec.ValuesOverride.V)
			if errInstallingChart != nil {
				return ctrl.Result{}, errInstallingChart
			}
			// force requeue to get new release state
			return ctrl.Result{Requeue: true}, nil
		}
	}

	if isReleasePending(releaseInfo) {
		r.Log.Info(fmt.Sprintf("%v is still in '%v' phase, checking back in a few..", req.NamespacedName, releaseInfo.Info.Status))
		return ctrl.Result{Requeue: true, RequeueAfter: 30 * time.Second}, nil
	}

	releaseValuesOverride := releaseInfo.Config
	if releaseValuesOverride == nil {
		releaseValuesOverride = map[string]interface{}{}
	}

	valuesInSync := reflect.DeepEqual(cr.Spec.ValuesOverride.V, releaseValuesOverride)
	chartVersionInSync := cr.Spec.Version == releaseInfo.Chart.Metadata.Version
	chartNameInSync := justChartName == releaseInfo.Chart.Metadata.Name

	if !chartNameInSync || !chartVersionInSync || !valuesInSync {
		r.Log.Info(fmt.Sprintf("%v release values in sync with installed values: %v", req.NamespacedName, valuesInSync))
		r.Log.Info(fmt.Sprintf("%v release chart version in sync with installed chart version: %v", req.NamespacedName, chartVersionInSync))
		r.Log.Info(fmt.Sprintf("%v release chart name in sync with installed chart name: %v", req.NamespacedName, chartNameInSync))

		chartPath, errPullingChart := r.pullChart(hrNamespace, hrName, repoAlias, chartName, cr.Spec.Version, helmV3)
		if errPullingChart != nil {
			if _, ok := errPullingChart.(pkg.ErrorHelmRepoNeedsRefresh); ok {
				r.Log.Info(fmt.Sprintf("refreshing helm repo index"))
				return ctrl.Result{Requeue: true}, helmV3.RefreshRepoIndex(repoAlias)
			}
			return ctrl.Result{}, errPullingChart
		}
		defer os.RemoveAll(strings.Split(chartPath, "/")[0])

		upgradeOpts := v3.UpgradeOptions{
			Namespace:                hrNamespace,
			DryRun:                   cr.Spec.DryRun,
			Wait:                     cr.Spec.Wait,
			Timeout:                  time.Duration(cr.Spec.WaitTimeout),
			ReleaseName:              hrName,
			DisableHooks:             cr.Spec.DisableHooks,
			DisableOpenAPIValidation: cr.Spec.DisableOpenAPIValidation,
			Atomic:                   cr.Spec.Atomic,
			CleanupOnFail:            cr.Spec.CleanupOnFail,
			SkipCRDs:                 !cr.Spec.IncludeCRDs,
			Force:                    cr.Spec.ForceUpgrade,
		}
		if _, errUpgradingRelease := helmV3.UpgradeRelease(chartPath, upgradeOpts, cr.Spec.ValuesOverride.V); errUpgradingRelease != nil {
			return ctrl.Result{}, errUpgradingRelease
		}
		r.Log.Info(fmt.Sprintf("Successfully upgraded helm release for %v", req.NamespacedName))
	}

	return ctrl.Result{}, nil
}
