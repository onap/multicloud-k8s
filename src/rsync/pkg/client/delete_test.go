package client

import (
	"fmt"
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClient_Delete(t *testing.T) {
	envContext := os.Getenv(contextEnvVarName)
	envKubeconfig := os.Getenv(kubeconfigEnvVarName)

	tests := []struct {
		name         string
		content      []byte
		exists       func(*Client) (string, error)
		wantExistErr bool
		context      string
		kubeconfig   string
		wantDelErr   bool
	}{
		{
			"delete non-existing resource",
			[]byte(`{"apiVersion": "v1", "kind": "ConfigMap", "metadata": { "name": "test-delete-0" }, "data": {	"key1": "apple" } }`),
			func(c *Client) (string, error) {
				cm, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-delete-0", metav1.GetOptions{})
				if err != nil {
					return "", err
				}
				if cm == nil {
					return "", fmt.Errorf("ConfigMap not found")
				}

				v, ok := cm.Data["key1"]
				if !ok {
					return "", fmt.Errorf("Data key not found")
				}

				return "key1 = " + v, nil
			},
			true,
			envContext, envKubeconfig,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewE(tt.context, tt.kubeconfig)
			if err != nil {
				t.Fatalf("failed to create the client with context %q and kubeconfig %q", tt.context, tt.kubeconfig)
			}
			if err := c.Delete(tt.content); (err != nil) != tt.wantDelErr {
				t.Errorf("Client.Delete() error = %v, wantErr %v", err, tt.wantDelErr)
			}

			if _, err := tt.exists(c); (err != nil) != tt.wantExistErr {
				t.Errorf("Delete test check failed. error = %v, wantErr %v", err, tt.wantExistErr)
				return
			}
		})
	}
}
