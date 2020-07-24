package utils

import (
	"context"
	"coveros.com/api/v1alpha1"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

func UnMarshalStringDataToHelmRelease(strData string) (*v1alpha1.HelmRelease, error) {
	helmRelease := &v1alpha1.HelmRelease{}
	if err := yaml.Unmarshal([]byte(strData), helmRelease); err != nil {
		return nil, err
	}

	if helmRelease.Kind == "" {
		return nil, nil
	}

	return helmRelease, nil
}

func RemoveDupesFromSlice(fromSlice []string) []string {
	m := make(map[string]int)
	var finalSlice []string

	for _, e := range fromSlice {
		m[e] = 1
	}

	for k, _ := range m {
		finalSlice = append(finalSlice, k)
	}
	return finalSlice
}

func SliceContainsString(slice []string, lookup string) (bool, int) {
	var contains bool
	var idx int = -1
	for i, e := range slice {
		if e == lookup {
			contains = true
			idx = i
			break
		}
	}
	return contains, idx
}

func UpdateCr(runtimeObj runtime.Object, client client.Client) error {
	return client.Update(context.TODO(), runtimeObj)
}

func UpdateCrStatus(runtimeObj runtime.Object, client2 client.Client) error {
	return client2.Status().Update(context.TODO(), runtimeObj)
}

func AddFinalizer(whichFinalizer string, client client.Client, cr *v1alpha1.HelmRelease) error {
	controllerutil.AddFinalizer(cr, whichFinalizer)
	return UpdateCr(cr, client)
}

func RemoveFinalizer(whichFinalizer string, client client.Client, cr *v1alpha1.HelmRelease) error {
	controllerutil.RemoveFinalizer(cr, whichFinalizer)
	return UpdateCr(cr, client)
}

func CreateNamespace(name string, client client.Client) error {
	ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name}, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := client.Create(context.TODO(), ns); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

func CreateHelmRelease(hr *v1alpha1.HelmRelease, client client.Client) (*v1alpha1.HelmRelease, error) {
	if hr.GetNamespace() == "" {
		hr.SetNamespace("default")
	}

	hrFound := &v1alpha1.HelmRelease{}
	err := client.Get(context.TODO(), types.NamespacedName{
		Namespace: hr.GetNamespace(),
		Name:      hr.GetName(),
	}, hrFound)
	if err != nil {
		if errors.IsNotFound(err) {
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
