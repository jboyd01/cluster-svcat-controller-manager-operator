package e2e

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/openshift/cluster-svcat-controller-manager-operator/pkg/operator"
	testframework "github.com/openshift/cluster-svcat-controller-manager-operator/test/framework"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Each of these tests helpers are similar, they only vary in the
// resource they are GETting and PATCHing.
// After the patch is done the test will poll the given resource.
// In case the operator is Managed state the patched data should
// not be equal to the one obtained after patch is applied.
// In case the operator is Unmanaged state the patched data should
// be equal to the one obtained after patch is applied.

var pollTimeout = 20 * time.Second

func patchAndCheckService(t *testing.T, client *testframework.Clientset, isOperatorManaged bool) error {
	t.Logf("patching Annotation on the target Service")
	service, err := client.Services(operator.OperandNamespace).Patch(operator.OperandServiceName, types.MergePatchType, []byte(`{"metadata": {"annotations": {"service.alpha.openshift.io/serving-cert-secret-name": "test"}}}`))
	if err != nil {
		return err
	}
	patchedData := service.GetAnnotations()

	if isOperatorManaged {
		t.Logf("polling for patched Annotation on the target Service to revert back to original data")
		err = wait.Poll(1*time.Second, pollTimeout, func() (stop bool, err error) {
			service, err = testframework.GetService(client)
			if err != nil {
				return true, err
			}
			newData := service.GetAnnotations()
			return !reflect.DeepEqual(patchedData, newData), nil
		})
		return err
	} else {
		t.Logf("polling for patched Annotation to ensure it doesn't revert")
		err = wait.Poll(1*time.Second, pollTimeout, func() (stop bool, err error) {
			service, err = testframework.GetService(client)
			if err != nil {
				return true, err
			}
			newData := service.GetAnnotations()
			if !reflect.DeepEqual(patchedData, newData) {
				return true, fmt.Errorf("annotation was reverted")
			}
			return false, nil
		})
		if err != nil && err == wait.ErrWaitTimeout {
			return nil
		}
		return err
	}
}
