package utils

import (
	"context"
	"errors"
	"github.com/coveros/genoa/api/v1alpha1"
	cNotifyLib "github.com/coveros/notification-library"
	"io"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)


func UpdateCr(runtimeObj runtime.Object, client client.Client) error {
	return client.Update(context.TODO(), runtimeObj)
}

func UpdateCrStatus(runtimeObj runtime.Object, client2 client.Client) error {
	return client2.Status().Update(context.TODO(), runtimeObj)
}

func AddFinalizer(whichFinalizer string, client client.Client, cr *v1alpha1.Release) error {
	controllerutil.AddFinalizer(cr, whichFinalizer)
	return UpdateCr(cr, client)
}

func RemoveFinalizer(whichFinalizer string, client client.Client, cr *v1alpha1.Release) error {
	controllerutil.RemoveFinalizer(cr, whichFinalizer)
	return UpdateCr(cr, client)
}

func CreateNamespace(name string, client client.Client) error {
	ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name}, ns)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			if err := client.Create(context.TODO(), ns); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

func CreateRelease(hr *v1alpha1.Release, client client.Client) (*v1alpha1.Release, error) {
	if hr.GetNamespace() == "" {
		hr.SetNamespace("default")
	}

	hrFound := &v1alpha1.Release{}
	err := client.Get(context.TODO(), types.NamespacedName{
		Namespace: hr.GetNamespace(),
		Name:      hr.GetName(),
	}, hrFound)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return hr, client.Create(context.TODO(), hr)
		}
		return nil, err
	}

	return hrFound, nil

}

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func DownloadFile(filepath, url, username, password string) (err error) {

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("StatusCodeNot200")
	}

	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err

}

func GetChannelIDForNotification(runtimeObjMeta metav1.ObjectMeta) string {
	channelToNotify := os.Getenv("DEFAULT_CHANNEL_ID")
	if channelID, ok := runtimeObjMeta.Annotations[SlackChannelIDAnnotation]; ok {
		channelToNotify = channelID
	}
	return channelToNotify
}

func getNotificationProvider() cNotifyLib.NotificationProvider {
	notificationProvider := cNotifyLib.Noop
	if val, ok := os.LookupEnv(EnvVarNotificationProvider); ok {
		notificationProvider = cNotifyLib.NotificationProvider(val)
	}
	return notificationProvider
}

func getNotificationProviderToken() string {
	notificationProviderToken := ""
	if val, ok := os.LookupEnv(EnvVarNotificationProviderToken); ok {
		notificationProviderToken = val
	}
	return notificationProviderToken
}

func NewNotifier() cNotifyLib.Notify {
	return cNotifyLib.NewNotificationProvider(getNotificationProvider(), getNotificationProviderToken())
}
