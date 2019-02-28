package framework

import (
	"fmt"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	operatorsv1 "github.com/openshift/api/operator/v1"
)

func isOperatorManaged(cr *operatorsv1.ServiceCatalogControllerManager) bool {
	return cr.Spec.ManagementState == operatorsv1.Managed
}

func isOperatorUnmanaged(cr *operatorsv1.ServiceCatalogControllerManager) bool {
	return cr.Spec.ManagementState == operatorsv1.Unmanaged
}

func isOperatorRemoved(cr *operatorsv1.ServiceCatalogControllerManager) bool {
	return cr.Spec.ManagementState == operatorsv1.Removed
}

type operatorStateReactionFn func(cr *operatorsv1.ServiceCatalogControllerManager) bool

func ensureServiceCatalogControllerManagerIsInDesiredState(t *testing.T, client *Clientset, state operatorsv1.ManagementState) error {
	var operatorConfig *operatorsv1.ServiceCatalogControllerManager
	// var checkFunc func()
	var checkFunc operatorStateReactionFn

	switch state {
	case operatorsv1.Managed:
		checkFunc = isOperatorManaged
	case operatorsv1.Unmanaged:
		checkFunc = isOperatorUnmanaged
	case operatorsv1.Removed:
		checkFunc = isOperatorRemoved
	}

	err := wait.Poll(1*time.Second, AsyncOperationTimeout, func() (stop bool, err error) {
		operatorConfig, err = client.ServiceCatalogControllerManagers().Get("cluster", metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return checkFunc(operatorConfig), nil
	})
	if err != nil {
		DumpObject(t, "the latest observed state of the service catalog resource", operatorConfig)
		DumpOperatorLogs(t, client)
		return fmt.Errorf("failed to wait to change service catalog operator state to 'Removed': %s", err)
	}

	if state == operatorsv1.Managed {
		errChan := make(chan error)
		go IsResourceAvailable(errChan, client, "Service")
		go IsResourceAvailable(errChan, client, "DaemonSet")
		checkErr := <-errChan
		if checkErr != nil {
			return fmt.Errorf("Failing waiting for resoruces to be managed: %v", checkErr)
		}
	} else if state == operatorsv1.Removed {
		err := WaitForResourceToNotExist(client, "DaemonSet")
		if err == nil {
			err = WaitForResourceToNotExist(client, "Service")
		}
		if err != nil {
			return fmt.Errorf("Failing wiating for resources to be Removed: %v", err)
		}
	}

	return nil
}

func ManageServiceCatalogControllerManager(t *testing.T, client *Clientset) error {
	cr, err := client.ServiceCatalogControllerManagers().Get("cluster", metav1.GetOptions{})
	if err != nil {
		return err
	}

	if isOperatorManaged(cr) {
		t.Logf("service catalog operator already in 'Managed' state")
		return nil
	}

	t.Logf("changing service catalog operator state to 'Managed'...")

	_, err = client.ServiceCatalogControllerManagers().Patch("cluster", types.MergePatchType, []byte(`{"spec": {"managementState": "Managed"}}`))
	if err != nil {
		return err
	}
	if err := ensureServiceCatalogControllerManagerIsInDesiredState(t, client, operatorsv1.Managed); err != nil {
		return fmt.Errorf("unable to change service catalog operator state to 'Managed': %s", err)
	}

	return nil
}

func UnmanageServiceCatalogControllerManager(t *testing.T, client *Clientset) error {
	cr, err := client.ServiceCatalogControllerManagers().Get("cluster", metav1.GetOptions{})
	if err != nil {
		return err
	}

	if isOperatorUnmanaged(cr) {
		t.Logf("service catalog operator already in 'Unmanaged' state")
		return nil
	}

	t.Logf("changing service catalog operator state to 'Unmanaged'...")

	_, err = client.ServiceCatalogControllerManagers().Patch("cluster", types.MergePatchType, []byte(`{"spec": {"managementState": "Unmanaged"}}`))
	if err != nil {
		return err
	}
	if err := ensureServiceCatalogControllerManagerIsInDesiredState(t, client, operatorsv1.Unmanaged); err != nil {
		return fmt.Errorf("unable to change service catalog operator state to 'Unmanaged': %s", err)
	}

	return nil
}

func RemoveServiceCatalogControllerManager(t *testing.T, client *Clientset) error {
	cr, err := client.ServiceCatalogControllerManagers().Get("cluster", metav1.GetOptions{})
	if err != nil {
		return err
	}

	if isOperatorRemoved(cr) {
		t.Logf("service catalog operator already in 'Removed' state")
		return nil
	}

	t.Logf("changing service catalog operator state to 'Removed'...")
	_, err = client.ServiceCatalogControllerManagers().Patch("cluster", types.MergePatchType, []byte(`{"spec": {"managementState": "Removed"}}`))
	if err != nil {
		return err
	}
	if err := ensureServiceCatalogControllerManagerIsInDesiredState(t, client, operatorsv1.Removed); err != nil {
		return fmt.Errorf("unable to change service catalog operator state to 'Removed': %s", err)
	}

	return nil
}
func MustManageServiceCatalogControllerManager(t *testing.T, client *Clientset) error {
	if err := ManageServiceCatalogControllerManager(t, client); err != nil {
		t.Fatal(err)
	}
	return nil
}

func MustUnmanageServiceCatalogControllerManager(t *testing.T, client *Clientset) error {
	if err := UnmanageServiceCatalogControllerManager(t, client); err != nil {
		t.Fatal(err)
	}
	return nil
}

func MustRemoveServiceCatalogControllerManager(t *testing.T, client *Clientset) error {
	if err := RemoveServiceCatalogControllerManager(t, client); err != nil {
		t.Fatal(err)
	}
	return nil
}
