package client

import (
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClient_Create_thenDelete(t *testing.T) {
	envContext := os.Getenv(contextEnvVarName)
	envKubeconfig := os.Getenv(kubeconfigEnvVarName)

	tests := []struct {
		name         string
		content      []byte
		checkIsThere func(*Client) (bool, error)
		wantItThere  bool
		context      string
		kubeconfig   string
		wantErr      bool
	}{
		{"create configMap", testData["create/cm.yaml"],
			func(c *Client) (bool, error) {
				cm, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-create-0", metav1.GetOptions{})
				if err != nil {
					return false, err
				}

				isThere := (cm.Data["key1"] == "apple")
				return isThere, nil
			}, true,
			envContext, envKubeconfig, false},
		{"create configMapList", testData["create/cml.yaml"],
			func(c *Client) (bool, error) {
				cm0, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-create-1", metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				cm1, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-create-2", metav1.GetOptions{})
				if err != nil {
					return false, err
				}

				isThere := (cm0.Data["key1"] == "strawberry" && cm1.Data["key2"] == "pinaple")
				return isThere, nil
			}, true,
			envContext, envKubeconfig, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewE(tt.context, tt.kubeconfig)
			if err != nil {
				t.Fatalf("failed to create the client with context %q and kubeconfig %q", tt.context, tt.kubeconfig)
			}

			if err := c.Create(tt.content); (err != nil) != tt.wantErr {
				t.Errorf("Client.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			isThere, err := tt.checkIsThere(c)
			if err != nil {
				t.Errorf("Client.Create() failed to check the applied resource. error = %v", err)
			} else if isThere != tt.wantItThere {
				t.Errorf("Client.Create() resource applied are not there")
			}

			if err := c.Delete(tt.content); err != nil {
				t.Errorf("Client.Delete() error = %v", err)
			}
		})
	}
}

func TestClient_CreateFile_thenDlete(t *testing.T) {
	envContext := os.Getenv(contextEnvVarName)
	envKubeconfig := os.Getenv(kubeconfigEnvVarName)

	tests := []struct {
		name         string
		filenames    []string
		checkIsThere func(*Client) (bool, error)
		wantItThere  bool
		context      string
		kubeconfig   string
		wantErr      bool
	}{
		{"create 1 configMap file", []string{"./testdata/create/cm.yaml"},
			func(c *Client) (bool, error) {
				cm, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-create-0", metav1.GetOptions{})
				if err != nil {
					return false, err
				}

				isThere := (cm.Data["key1"] == "apple")
				return isThere, nil
			}, true,
			envContext, envKubeconfig, false},
		{"ceate 2 configMap files", []string{"./testdata/create/cm.yaml", "./testdata/create/cml.yaml"},
			func(c *Client) (bool, error) {
				cm0, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-create-1", metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				cm1, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-create-2", metav1.GetOptions{})
				if err != nil {
					return false, err
				}

				isThere := (cm0.Data["key1"] == "strawberry" && cm1.Data["key2"] == "pinaple")
				return isThere, nil
			}, true,
			envContext, envKubeconfig, false},
		{"create secret from URL", []string{"https://raw.githubusercontent.com/johandry/klient/master/testdata/create/secret.yaml"},
			func(c *Client) (bool, error) {
				s, err := c.Clientset.CoreV1().Secrets("default").Get("test-secret-create-0", metav1.GetOptions{})
				if err != nil {
					return false, err
				}

				isThere := (string(s.Data["password"]) == "Super5ecret0!")
				return isThere, nil
			}, true,
			envContext, envKubeconfig, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewE(tt.context, tt.kubeconfig)
			if err != nil {
				t.Fatalf("failed to create the client with context %q and kubeconfig %q", tt.context, tt.kubeconfig)
			}

			if err := c.CreateFile(tt.filenames...); (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			isThere, err := tt.checkIsThere(c)
			if err != nil {
				t.Errorf("Client.CreateFile() failed to check the applied resource. error = %v", err)
			} else if isThere != tt.wantItThere {
				t.Errorf("Client.CreateFile() resource applied are not there")
			}

			if err := c.DeleteFiles(tt.filenames...); err != nil {
				t.Errorf("Client.Delete() error = %v", err)
			}
		})
	}
}
