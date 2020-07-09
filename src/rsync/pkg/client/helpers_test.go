package client

import (
	"log"
	"os"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	contextEnvVarName    = "KUBECLIENT_TEST_CONTEXT"
	kubeconfigEnvVarName = "KUBECLIENT_TEST_KUBECONFIG"
	versionEnvVarName    = "KUBECLIENT_TEST_VERSION"
)

func init() {
	envContext := os.Getenv(contextEnvVarName)
	envKubeconfig := os.Getenv(kubeconfigEnvVarName)
	c, err := NewE(envContext, envKubeconfig)
	if err != nil {
		log.Println("You may not have a kubernetes cluster properly configured to run the tests. Create a cluster either with Kind or Docker Desktop to execute the tests and make sure it's configured correctly")
		log.Fatalf("failed to create the client with context %q and kubeconfig %q", envContext, envKubeconfig)
	}

	if _, err := c.Version(); err != nil {
		log.Println("You may not have a kubernetes cluster to run the tests. Create a cluster either with Kind or Docker Desktop to execute the tests")
		log.Fatalf("the Kubernetes cluster is not reachable")
	}
}

func TestClient_CreateAndDeleteNamespace(t *testing.T) {
	envContext := os.Getenv(contextEnvVarName)
	envKubeconfig := os.Getenv(kubeconfigEnvVarName)

	tests := []struct {
		name          string
		namespace     string
		context       string
		kubeconfig    string
		wantCreateErr bool
		wantDeleteErr bool
	}{
		{"namespace test 001", "test-ns-001", envContext, envKubeconfig, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewE(tt.context, tt.kubeconfig)
			if err != nil {
				t.Fatalf("failed to create the client with context %q and kubeconfig %q", tt.context, tt.kubeconfig)
			}
			if err := c.CreateNamespace(tt.namespace); (err != nil) != tt.wantCreateErr {
				t.Errorf("Client.CreateNamespace() error = %v, wantErr %v", err, tt.wantCreateErr)
			}

			if _, err := c.Clientset.CoreV1().Namespaces().Get(tt.namespace, metav1.GetOptions{}); err != nil {
				if errors.IsNotFound(err) {
					t.Errorf("Client.CreateNamespace() failed to create the namespace %q, it was not found. Error: %v", tt.namespace, err)
					return
				}
				t.Fatalf("failed to get the namespace %q. Error: %v", tt.namespace, err)
			}

			if err := c.DeleteNamespace(tt.namespace); (err != nil) != tt.wantDeleteErr {
				t.Errorf("Client.DeleteNamespace() error = %v, wantErr %v", err, tt.wantDeleteErr)
			}
		})
	}
}

func TestClient_Version(t *testing.T) {
	envContext := os.Getenv(contextEnvVarName)
	envKubeconfig := os.Getenv(kubeconfigEnvVarName)

	tests := []struct {
		name       string
		context    string
		kubeconfig string
		want       string // Use the exact version to expect or `>N.M` if you want a version greater than vM.N
		wantErr    bool
	}{
		{"unreachable", "unknown", "", "dosen't matter", true},
		// {"version 1.15.5", envContext, envKubeconfig, "v1.15.5", false},
		{"version greater than 1.14", envContext, envKubeconfig, ">1.14", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewE(tt.context, tt.kubeconfig)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			got, err := c.Version()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Version() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var pass bool
			switch tt.want[:1] {
			case "v":
				pass = (got == tt.want)
			case ">":
				wantVer := tt.want[1:] // remove the `>`
				wantV := strings.Split(wantVer, ".")
				wantMaxVer := wantV[0]
				wantMinVer := wantV[1]

				gotVer := got[1:] // remove the `v`
				gotV := strings.Split(gotVer, ".")
				gotMaxVer := gotV[0]
				gotMinVer := gotV[1]

				pass = ((gotMaxVer >= wantMaxVer) && (gotMinVer >= wantMinVer))
			default:
				t.Errorf("incorrect expected value. Fix test %q", tt.name)
				return
			}

			if !pass {
				t.Errorf("Client.Version() = %v, want %v", got, tt.want)
			}
		})
	}
}
