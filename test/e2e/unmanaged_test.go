package e2e

import (
	"testing"

	testframework "github.com/openshift/cluster-svcat-controller-manager-operator/test/framework"
)

// TestUnmanaged sets operator to Unmanaged state. After that Service Catalog Controller Manager
// daemonset is deleted after which the daemonset is tested for unavailability, to
// check that it wasn't recreated byt the operator. Other resources from the
// 'operator.OperandNamespace' namespace (Service) are tested for availability
// since they have not been deleted.
func TestUnmanaged(t *testing.T) {
	client := testframework.MustNewClientset(t, nil)
	defer testframework.MustManageServiceCatalogControllerManager(t, client)
	testframework.MustUnmanageServiceCatalogControllerManager(t, client)
	testframework.DeleteAll(t, client)

	t.Logf("verifying the operator does not recreate deleted resources...")
	errChan := make(chan error)
	go testframework.IsResourceUnavailable(errChan, client, "Service")
	go testframework.IsResourceUnavailable(errChan, client, "DaemonSet")
	checkErr := <-errChan

	if checkErr != nil {
		t.Fatal(checkErr)
	}
}

func TestEditUnmanagedService(t *testing.T) {
	client := testframework.MustNewClientset(t, nil)
	defer testframework.MustManageServiceCatalogControllerManager(t, client)
	testframework.MustUnmanageServiceCatalogControllerManager(t, client)

	err := patchAndCheckService(t, client, false)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
}
