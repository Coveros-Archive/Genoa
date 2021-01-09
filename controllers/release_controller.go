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
	"errors"
	"fmt"
	"github.com/coveros/genoa/pkg"
	v3 "github.com/coveros/genoa/pkg/helm/v3"
	"github.com/coveros/genoa/pkg/utils"
	cNotifyLib "github.com/coveros/notification-library"
	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/storage/driver"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"os"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
	"time"

	coverosv1alpha1 "github.com/coveros/genoa/api/v1alpha1"
)

// ReleaseReconciler reconciles a Release object
type ReleaseReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Cfg      *rest.Config
	Notifier cNotifyLib.Notify
}

func (r *ReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: 7}).
		For(&coverosv1alpha1.Release{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
			},
		}).
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
	notificationChannel := utils.GetChannelIDForNotification(cr.ObjectMeta)
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
		r.Notifier.SendMsg(cNotifyLib.NotifyTemplate{
			Channel:   notificationChannel,
			Title:     req.NamespacedName.String(),
			EventType: cNotifyLib.Success,
			Fields: map[string]string{
				"Chart":     cr.Spec.Chart + "-" + cr.Spec.Version,
				"Namespace": cr.GetNamespace(),
				"Reason":    "Release deleted successfully :boom:"},
		})
		return ctrl.Result{}, nil // do not requeue
	}

	if cr.Spec.DependsOn.GetName() != "" {
		name := cr.Spec.DependsOn.GetName()
		ns := cr.Spec.DependsOn.GetNamespace()
		if ns == "" {
			ns = "default"
		}

		dependsOnCr := &coverosv1alpha1.Release{}
		if errGettingDependsOnCr := r.Client.Get(context.TODO(),
			types.NamespacedName{Name: name, Namespace: ns}, dependsOnCr); errGettingDependsOnCr != nil {
			return ctrl.Result{}, errGettingDependsOnCr
		}
		// wait until the parent release is installed
		if !dependsOnCr.Status.Installed {
			r.Log.Info(fmt.Sprintf("%v depends on %v/%v and is not ready yet.. will re-check back in a few...", req.NamespacedName, ns, name))
			return ctrl.Result{Requeue: true, RequeueAfter: 60 * time.Second}, nil
		}
	}

	if (cr.Status.FailureCount >= cr.Spec.MaxRetries) && cr.Status.FailureCount != 0 {
		r.Log.Info(fmt.Sprintf("%v has reached max reconcile limit, please update spec.maxRetries if you want to retry", req.NamespacedName))
		return ctrl.Result{}, nil
	}

	releaseInfo, errGettingReleaseInfo := helmV3.GetRelease(hrName)
	if errGettingReleaseInfo != nil {
		if errors.Is(errGettingReleaseInfo, driver.ErrReleaseNotFound) {
			r.Log.Info("release not found, installing now...")

			chartPath, errPullingChart := r.pullChart(hrNamespace, hrName, repoAlias, chartName, cr.Spec.Version, helmV3)
			if errPullingChart != nil {
				if _, ok := errPullingChart.(pkg.ErrorHelmRepoNeedsRefresh); ok {
					cr.Status.FailureCount++
					if err := r.Client.Status().Update(context.TODO(), cr); err != nil {
						return ctrl.Result{}, err
					}
					return ctrl.Result{Requeue: true}, helmV3.RefreshRepoIndex(repoAlias)
				}
				return ctrl.Result{}, errPullingChart
			}
			defer os.RemoveAll(strings.Split(chartPath, "/")[0])
			r.Log.Info(fmt.Sprintf("%v: downloaded chart at %v", req.NamespacedName, chartPath))
			installOpts := getReleaseInstallOptions(cr)
			_, errInstallingChart := helmV3.InstallRelease(chartPath, installOpts, cr.Spec.ValuesOverride.V)
			if errInstallingChart != nil {
				r.Notifier.SendMsg(cNotifyLib.NotifyTemplate{
					Channel:   notificationChannel,
					Title:     req.NamespacedName.String(),
					EventType: cNotifyLib.Failure,
					Fields: map[string]string{
						"Chart":     cr.Spec.Chart + "-" + cr.Spec.Version,
						"Namespace": cr.GetNamespace(),
						"Reason":    fmt.Sprintf("Release failed to install :bug: :construction: %v", errInstallingChart)},
				})
				cr.Status.FailureCount++
				if err := r.Client.Status().Update(context.TODO(), cr); err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, errInstallingChart
			}
			// force requeue to get new release state
			r.Notifier.SendMsg(cNotifyLib.NotifyTemplate{
				Channel:   notificationChannel,
				Title:     req.NamespacedName.String(),
				EventType: cNotifyLib.Success,
				Fields: map[string]string{
					"Chart":     cr.Spec.Chart + "-" + cr.Spec.Version,
					"Namespace": cr.GetNamespace(),
					"Reason":    "Release installed successfully :smile:"},
			})
			cr.Status.Installed = true
			cr.Status.FailureCount = 0
			if err := utils.UpdateCrStatus(cr, r.Client); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, r.Client.Status().Update(context.TODO(), cr)
		}
		return ctrl.Result{}, errGettingReleaseInfo
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
	//releaseRevisionInSync := cr.Status.RevisionNumber == releaseInfo.Version

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

		// TODO: need to figure out what to when we get immutable error from a revision rollback
		/**
		Example: error
		Service is invalid: spec.clusterIP: Invalid value: "":
		field is immutable &&
		failed to replace object: Service "jenkins" is invalid: spec.clusterIP: Invalid value: "": field is immutable
		*/
		//if !releaseRevisionInSync {
		//	r.Log.Info(fmt.Sprintf("%v rolling back to revision number: %v", req.NamespacedName, cr.Status.RevisionNumber))
		//	newRevisionNumber := (releaseInfo.Version - cr.Status.RevisionNumber) + cr.Status.RevisionNumber + 1
		//	rollBackOpts := v3.RollbackToRevisionOptions{Force: true, ToRevision: cr.Status.RevisionNumber}
		//	if errRollingBack := helmV3.RollbackToRevision(hrName, rollBackOpts); errRollingBack != nil {
		//		return ctrl.Result{}, errRollingBack
		//	}
		//	r.Log.Info(fmt.Sprintf("%v successfully rolled back to revision number: %v", req.NamespacedName, cr.Status.RevisionNumber))
		//	cr.Status.RevisionNumber = newRevisionNumber
		//	r.Log.Info(fmt.Sprintf("%v new revision number: %v", req.NamespacedName, newRevisionNumber))
		//	return ctrl.Result{}, r.Client.Status().Update(context.TODO(), cr)
		//}
		upgradeOpts := getReleaseUpgradeOptions(cr)
		if _, errUpgradingRelease := helmV3.UpgradeRelease(chartPath, upgradeOpts, cr.Spec.ValuesOverride.V); errUpgradingRelease != nil {

			r.Notifier.SendMsg(cNotifyLib.NotifyTemplate{
				Channel:   notificationChannel,
				Title:     req.NamespacedName.String(),
				EventType: cNotifyLib.Failure,
				Fields: map[string]string{
					"Chart":     cr.Spec.Chart + "-" + cr.Spec.Version,
					"Namespace": cr.GetNamespace(),
					"Reason":    fmt.Sprintf("Release failed to upgrade :bug: :construction: %v", errUpgradingRelease)},
			})
			cr.Status.FailureCount++
			if err := r.Client.Status().Update(context.TODO(), cr); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, errUpgradingRelease
		}
		r.Notifier.SendMsg(cNotifyLib.NotifyTemplate{
			Channel:   notificationChannel,
			Title:     req.NamespacedName.String(),
			EventType: cNotifyLib.Success,
			Fields: map[string]string{
				"Chart":     cr.Spec.Chart + "-" + cr.Spec.Version,
				"Namespace": cr.GetNamespace(),
				"Reason":    "Release upgraded successfully :confetti_ball:"},
		})
		r.Log.Info(fmt.Sprintf("Successfully upgraded helm release for %v", req.NamespacedName))
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}
