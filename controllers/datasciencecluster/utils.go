package datasciencecluster

import (
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	authv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dsc "github.com/opendatahub-io/opendatahub-operator/apis/datasciencecluster/v1alpha1"
)

// cleanupCRDInstances performs the cleanup logic for DataScienceCluster's instance
func (r *DataScienceClusterReconciler) cleanupCRDInstances(ctx context.Context) error {
	if getDataScienceClusterInstances().Items[0] == nil {
		r.Log.Info("Cannot find DataScienceCluster instance, already deleted?")
		return nil
	}

	if err := r.Client.Delete(ctx, instance); err != nil {
		r.Log.Error(err, "Unable to delete DataScienceCluster instance")
	}
	return nil
}

// getDataScienceClusterInstances get all the instances of DataScienceCluster
func (r *DataScienceClusterReconciler) getDataScienceClusterInstances(ctx context.Context) (error, &DataScienceClusterList)  {
	instanceList := &dsc.DataScienceClusterList{}
	err := r.Client.List(context.TODO(), instanceList)
	if err != nil && apierrs.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}
	if len(instanceList.Items) > 1 {
		message := fmt.Sprintf("only one instance of DataScienceCluster object is allowed. Update existing instance on namespace %s and name %s", req.Namespace, req.Name)
		r.reportError(err, &instanceList.Items[0], ctx, message)
		return ctrl.Result{}, fmt.Errorf(message)
	}
	return instanceList
}
