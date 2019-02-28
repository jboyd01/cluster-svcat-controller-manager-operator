package framework

import (
	"fmt"
	"testing"
	"time"

	"github.com/openshift/cluster-svcat-controller-manager-operator/pkg/operator"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	// AsyncOperationTimeout is how long we want to wait for asynchronous
	// operations to complete. ForeverTestTimeout is not long enough to create
	// several replicas and get them available on a slow machine.
	// Setting this to 5 minutes

	AsyncOperationTimeout = 5 * time.Minute
)

func DeleteAll(t *testing.T, client *Clientset) {
	resources := []string{"DaemonSets", "Service"}

	for _, resource := range resources {
		t.Logf("deleting service catalog %s...", resource)
		if err := DeleteCompletely(
			func() (runtime.Object, error) {
				return GetResource(client, resource)
			},
			func(*metav1.DeleteOptions) error {
				return deleteResource(client, resource)
			},
		); err != nil {
			t.Fatalf("unable to delete service catalog resource %s: %s", resource, err)
		}
	}
}

func GetResource(client *Clientset, resource string) (runtime.Object, error) {
	var res runtime.Object
	var err error
	switch resource {
	case "Service":
		res, err = GetService(client)
	case "DaemonSet":
		fallthrough
	default:
		res, err = GetDaemonSet(client)
	}
	return res, err
}

func GetService(client *Clientset) (*corev1.Service, error) {
	return client.Services(operator.OperandNamespace).Get(operator.OperandServiceName, metav1.GetOptions{})
}

func GetDaemonSet(client *Clientset) (*appv1.DaemonSet, error) {
	return client.DaemonSets(operator.OperandNamespace).Get(operator.OperandServiceName, metav1.GetOptions{})
}

func deleteResource(client *Clientset, resource string) error {
	var err error
	switch resource {
	case "Service":
		err = client.Services(operator.OperandNamespace).Delete(operator.OperandServiceName, &metav1.DeleteOptions{})
	case "DaemonSet":
		fallthrough
	default:
		err = client.DaemonSets(operator.OperandNamespace).Delete(operator.OperandServiceName, &metav1.DeleteOptions{})
	}
	return err
}

// DeleteCompletely sends a delete request and waits until the resource and
// its dependents are deleted.
func DeleteCompletely(getObject func() (runtime.Object, error), deleteObject func(*metav1.DeleteOptions) error) error {
	obj, err := getObject()
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	accessor, _ := meta.Accessor(obj)
	uid := accessor.GetUID()

	policy := metav1.DeletePropagationForeground
	if err := deleteObject(&metav1.DeleteOptions{
		Preconditions: &metav1.Preconditions{
			UID: &uid,
		},
		PropagationPolicy: &policy,
	}); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return wait.Poll(1*time.Second, AsyncOperationTimeout, func() (stop bool, err error) {
		obj, err = getObject()
		if err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}

		accessor, _ := meta.Accessor(obj)

		return accessor.GetUID() != uid, nil
	})
}

// IsResourceAvailable checks if tested resource is available (recreated by service catalog operator)
// during 45 second period. If resource is not found an error will be returned.
func IsResourceAvailable(errChan chan error, client *Clientset, resource string) {
	counter := 0
	err := wait.Poll(1*time.Second, 60*time.Second, func() (stop bool, err error) {
		_, err = GetResource(client, resource)
		if err == nil {
			return true, nil
		}
		if counter == 45 {
			if err != nil {
				return true, fmt.Errorf("deleted service catalog resource %s was not recreated", resource)
			}
			return true, nil
		}
		counter++
		return false, nil
	})
	errChan <- err
}

// Wait for up to 45 seconds checking if the resource has been deleted.  Returns
// non nil error on timeout or other error
func WaitForResourceToNotExist(client *Clientset, resource string) error {
	return wait.PollImmediate(1*time.Second, 45*time.Second,
		func() (bool, error) {
			_, err := GetResource(client, resource)
			if nil == err {
				return false, nil
			}

			if errors.IsNotFound(err) {
				return true, nil
			}

			return false, nil
		},
	)
}

// IsResourceUnavailable checks if tested resource is unavailable(not recreated by service catalog-operator)
// If not error will be returned.
func IsResourceUnavailable(errChan chan error, client *Clientset, resource string) {
	counter := 0
	err := wait.Poll(1*time.Second, 45*time.Second, func() (stop bool, err error) {
		_, err = GetResource(client, resource)
		if err == nil {
			return true, fmt.Errorf("deleted service catalog %s was recreated\n", resource)
		}
		if !errors.IsNotFound(err) {
			return true, err
		}
		counter++
		if counter == 45 {
			return true, nil
		}
		return false, nil
	})
	errChan <- err
}
