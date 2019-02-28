package e2e

import (
	"testing"

	testframework "github.com/openshift/cluster-svcat-controller-manager-operator/test/framework"
)

// TestManaged sets operator to Managed state. After that service catalog controller manager
// daemonset is deleted after which the daemonset is tested for availability, together
// with other resources from the 'operator.OperandNamespace' namespace
func TestManaged(t *testing.T) {
	client := testframework.MustNewClientset(t, nil)
	defer testframework.MustManageServiceCatalogControllerManager(t, client)
	testframework.MustManageServiceCatalogControllerManager(t, client)

	t.Logf("deleting the deployment")
	testframework.DeleteAll(t, client)
	t.Logf("waiting for resources to not exist")
	err := testframework.WaitForResourceToNotExist(client, "DaemonSet")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("verifying the operator recreates the deployment...")

	errChan := make(chan error)
	go testframework.IsResourceAvailable(errChan, client, "Service")
	go testframework.IsResourceAvailable(errChan, client, "DaemonSet")
	checkErr := <-errChan

	if checkErr != nil {
		t.Fatal(checkErr)
	}
}

func TestEditManagedService(t *testing.T) {
	client := testframework.MustNewClientset(t, nil)
	defer testframework.MustManageServiceCatalogControllerManager(t, client)
	testframework.MustManageServiceCatalogControllerManager(t, client)

	err := patchAndCheckService(t, client, true)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
}
