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
	v3 "coveros.com/pkg/helm/v3"
	"coveros.com/pkg/utils"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"strings"
	"time"

	coverosv1alpha1 "coveros.com/api/v1alpha1"
)

// HelmReleaseReconciler reconciles a HelmRelease object
type HelmReleaseReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Cfg    *rest.Config
}

const (
	ReleaseFinalizer              = "agill.apps.helmRelease"
	AutoDeleteNamespaceAnnotation = ReleaseFinalizer + "/autoDeleteNamespace"
)

// +kubebuilder:rbac:groups=coveros.coveros.com,resources=helmreleases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=coveros.coveros.com,resources=helmreleases/status,verbs=get;update;patch
func (r *HelmReleaseReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("helmrelease", req.NamespacedName)

	// Fetch the HelmRelease CR
	cr := &coverosv1alpha1.HelmRelease{}
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
	repoWithChartName := strings.Split(cr.Spec.Chart, "/")
	repoAlias, chartName := repoWithChartName[0], repoWithChartName[1]
	actionConfig, errCreatingActionConfig := v3.NewActionConfig(cr.GetNamespace(), r.Cfg)
	if errCreatingActionConfig != nil {
		return ctrl.Result{}, errCreatingActionConfig
	}

	// add finalizer
	if errAddingFinalizer := utils.AddFinalizer(ReleaseFinalizer, r.Client, cr); errAddingFinalizer != nil {
		return ctrl.Result{}, errAddingFinalizer
	}

	// handle delete
	if cr.GetDeletionTimestamp() != nil {
		if errCleaningUp := r.cleanup(cr, actionConfig); errCleaningUp != nil {
			return ctrl.Result{}, errCleaningUp
		}
		return ctrl.Result{}, nil // do not requeue
	}

	releaseInfo, errGettingReleaseInfo := actionConfig.GetRelease(hrName)
	if errGettingReleaseInfo != nil {
		if errors.Is(errGettingReleaseInfo, driver.ErrReleaseNotFound) {
			r.Log.Info("release not found, installing now...")

			chartPath, errPullingChart := r.pullChart(hrNamespace, hrName, repoAlias, chartName, cr.Spec.Version, actionConfig)
			if errPullingChart != nil {
				return ctrl.Result{}, errPullingChart
			}

			installOptions := v3.InstallOptions{
				Namespace:   hrNamespace,
				DryRun:      cr.Spec.DryRun,
				Wait:        cr.Spec.Wait,
				Timeout:     time.Duration(cr.Spec.WaitTimeout),
				ReleaseName: hrName,
			}
			r.Log.Info(fmt.Sprintf("%v: downloaded chart at %v", req.NamespacedName, chartPath))
			_, errInstallingChart := actionConfig.InstallRelease(chartPath, installOptions, cr.Spec.ValuesOverride.V)
			if errInstallingChart != nil {
				return ctrl.Result{}, errInstallingChart
			}
			// force requeue to get new release state
			return ctrl.Result{Requeue: true}, nil
		}
	}

	if releaseInfo.Info.Status == release.StatusPendingInstall ||
		releaseInfo.Info.Status == release.StatusUninstalling ||
		releaseInfo.Info.Status == release.StatusPendingRollback ||
		releaseInfo.Info.Status == release.StatusPendingUpgrade {

		r.Log.Info(fmt.Sprintf("%v is still in '%v' phase, checking back in a few..", req.NamespacedName, releaseInfo.Info.Status))
		return ctrl.Result{Requeue: true, RequeueAfter: 30 * time.Second}, nil
	}

	releaseValuesOverride := releaseInfo.Config
	if releaseValuesOverride == nil {
		releaseValuesOverride = map[string]interface{}{}
	}

	valuesInSync := reflect.DeepEqual(cr.Spec.ValuesOverride.V, releaseValuesOverride)
	chartVersionInSync := cr.Spec.Version == releaseInfo.Chart.Metadata.Version
	chartNameInSync := cr.Spec.Chart == releaseInfo.Chart.Metadata.Name

	if !valuesInSync || !chartVersionInSync || !chartNameInSync {
		chartPath, errPullingChart := r.pullChart(hrNamespace, hrName, repoAlias, chartName, cr.Spec.Version, actionConfig)
		if errPullingChart != nil {
			return ctrl.Result{}, errPullingChart
		}
		upgradeOpts := v3.UpgradeOptions{
			Namespace:   hrNamespace,
			DryRun:      cr.Spec.DryRun,
			Wait:        cr.Spec.Wait,
			Timeout:     time.Duration(cr.Spec.WaitTimeout),
			ReleaseName: hrName,
		}
		if _, errUpgradingRelease := actionConfig.UpgradeRelease(chartPath, upgradeOpts, cr.Spec.ValuesOverride.V); errUpgradingRelease != nil {
			return ctrl.Result{}, errUpgradingRelease
		}
	}

	return ctrl.Result{}, nil
}

func (r *HelmReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: 5}).
		For(&coverosv1alpha1.HelmRelease{}).
		Complete(r)
}
