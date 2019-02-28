package e2e

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/openshift/cluster-svcat-controller-manager-operator/pkg/operator"
	testframework "github.com/openshift/cluster-svcat-controller-manager-operator/test/framework"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeset "k8s.io/client-go/kubernetes"
)

var (
	kubeClient *kubeset.Clientset
)

func TestMain(m *testing.M) {

	kubeconfig, err := testframework.GetConfig()
	if err != nil {
		fmt.Printf("unable to get kubeconfig: %s", err)
		os.Exit(1)
	}

	kubeClient, err = kubeset.NewForConfig(kubeconfig)
	if err != nil {
		fmt.Printf("%#v", err)
		os.Exit(1)
	}

	// e2e test job does not guarantee our operator is up before
	// launching the test, so we need to do so.
	fmt.Println("checking for operator availability")
	err = waitForOperator()
	if err != nil {
		fmt.Println("failed waiting for operator to start")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func waitForOperator() error {
	depClient := kubeClient.AppsV1().Deployments(operator.OperatorNamespace)
	err := wait.PollImmediate(1*time.Second, 10*time.Minute, func() (bool, error) {
		_, err := depClient.Get(operator.OperatorNamespace, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("error waiting for operator deployment to exist: %v\n", err)
			return false, nil
		}
		fmt.Println("found operator deployment")
		return true, nil
	})
	return err
}
