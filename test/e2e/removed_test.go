package e2e

import (
	"fmt"
	"testing"

	testframework "github.com/openshift/cluster-svcat-controller-manager-operator/test/framework"
)

// TestRemoved sets operator to Removed state. After that all the resources
// from the 'operator.OperandNamespace' namespace (DaemonSet, Service),
// are tested for unavailability since the operator should delete them.
func TestRemoved(t *testing.T) {
	client := testframework.MustNewClientset(t, nil)

	fmt.Println("Ensuring deployment exists...")
	errChan := make(chan error)
	go testframework.IsResourceAvailable(errChan, client, "Service")
	go testframework.IsResourceAvailable(errChan, client, "DaemonSet")
	checkErr := <-errChan

	if checkErr != nil {
		t.Fatal(checkErr)
	}

	fmt.Println("Setting to managedState=Removed...")

	defer testframework.MustManageServiceCatalogControllerManager(t, client)
	testframework.MustRemoveServiceCatalogControllerManager(t, client)

	fmt.Println("waiting to ensure Service Catalog resources are deleted...")

	err := testframework.WaitForResourceToNotExist(client, "DaemonSet")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Verifying the resources remain deleted")
	errChan = make(chan error)
	go testframework.IsResourceUnavailable(errChan, client, "Service")
	go testframework.IsResourceUnavailable(errChan, client, "DaemonSet")
	checkErr = <-errChan

	if checkErr != nil {
		t.Fatal(checkErr)
	}
}
