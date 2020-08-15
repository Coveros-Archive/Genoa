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

package main

import (
	v3 "coveros.com/pkg/helm/v3"
	"flag"
	"github.com/containrrr/shoutrrr/pkg/router"
	"net/url"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	coverosv1alpha1 "coveros.com/api/v1alpha1"
	"coveros.com/controllers"
	"github.com/containrrr/shoutrrr"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	SlackUrlEnvVar = "SLACK_WEBHOOK_URL"
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = coverosv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var customRepoConfigPath string
	var notificationChannels []string
	var notificationSender *router.ServiceRouter
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for release manager. "+
			"Enabling this will ensure there is only one active release manager.")
	flag.StringVar(&customRepoConfigPath, "custom-helm-repos-file", "", "Your own custom helm repo files")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	syncPeriod := 3 * time.Minute
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "a8248481.coveros.com",
		SyncPeriod:         &syncPeriod,
		Namespace:          "",
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if customRepoConfigPath != "" {
		if errAddingCustomRepos := v3.AddReposFromFile(customRepoConfigPath); errAddingCustomRepos != nil {
			setupLog.Error(errAddingCustomRepos, "Failed to add custom helm repos")
			os.Exit(1)
		}
	}

	releaseReconciler := &controllers.ReleaseReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("release"),
		Scheme: mgr.GetScheme(),
		Cfg:    mgr.GetConfig(),
	}

	// if slackUrl is set, add that to list of notification channels
	if slackUrl, ok := os.LookupEnv(SlackUrlEnvVar); ok {
		if _, errParsingUrl := url.ParseRequestURI(slackUrl); errParsingUrl != nil {
			setupLog.Error(errParsingUrl, "Failed to parse slack url! Exiting now...")
			os.Exit(1)
		}
		// we can simply keep on adding other alert urls to notificationChannels
		// as long as it is supported by shoutrrr : https://containrrr.dev/shoutrrr/services/overview/
		notificationChannels = append(notificationChannels, slackUrl)
	}

	// if notificationChannels has any urls, we need to setup a notification router
	if len(notificationChannels) > 0 {
		notificationSender, err = shoutrrr.CreateSender(notificationChannels...)
		if err != nil {
			setupLog.Error(err, "Failed to setup a notification sender.. exiting now..")
			os.Exit(1)
		}
		releaseReconciler.NotificationSender = notificationSender
	}

	if err = releaseReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "release")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
