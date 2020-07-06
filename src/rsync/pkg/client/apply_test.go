package client

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var testData map[string][]byte

func init() {
	testData = make(map[string][]byte)
	for _, f := range []string{"apply/cm.yaml", "create/cm.yaml", "apply/cml.yaml", "create/cml.yaml"} {
		content, err := ioutil.ReadFile("./testdata/" + f)
		if err != nil {
			log.Fatalf("cannot load test data from file %q", f)
		}
		// name := strings.Split(f,".")[0]
		testData[f] = content
	}
}

func TestClient_Apply_thenDelete(t *testing.T) {
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
		{"apply configMap", testData["apply/cm.yaml"],
			func(c *Client) (bool, error) {
				cm, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-apply-0", metav1.GetOptions{})
				if err != nil {
					return false, err
				}

				isThere := (cm.Data["key1"] == "apple")
				return isThere, nil
			}, true,
			envContext, envKubeconfig, false},
		{"apply configMapList", testData["apply/cml.yaml"],
			func(c *Client) (bool, error) {
				cm0, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-apply-1", metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				cm1, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-apply-2", metav1.GetOptions{})
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
			if err := c.Apply(tt.content); (err != nil) != tt.wantErr {
				t.Errorf("Client.Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			isThere, err := tt.checkIsThere(c)
			if err != nil {
				t.Errorf("Client.Apply() failed to check the applied resource. error = %v", err)
			} else if isThere != tt.wantItThere {
				t.Errorf("Client.Apply() resource applied are not there")
			}

			if err := c.Delete(tt.content); err != nil {
				t.Errorf("Client.Delete() error = %v", err)
			}
		})
	}
}

func TestClient_Apply_PatchAndDelete(t *testing.T) {
	envContext := os.Getenv(contextEnvVarName)
	envKubeconfig := os.Getenv(kubeconfigEnvVarName)

	tests := []struct {
		name           string
		initial        []byte
		modified       []byte
		getChange      func(*Client) (string, error)
		want           string
		context        string
		kubeconfig     string
		wantInitialErr bool
		wantPatchErr   bool
		wantGetErr     bool
	}{
		{
			"apply & patch configMap test-applypatch-0",
			[]byte(`{"apiVersion": "v1", "kind": "ConfigMap", "metadata": { "name": "test-applypatch-0" }, "data": {	"key1": "apple" } }`),
			[]byte(`{"apiVersion": "v1", "kind": "ConfigMap", "metadata": { "name": "test-applypatch-0" }, "data": {	"key1": "orange" } }`),
			func(c *Client) (string, error) {
				cm, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-applypatch-0", metav1.GetOptions{})
				if err != nil {
					return "", err
				}

				return "key1: " + cm.Data["key1"], nil
			},
			"key1: orange",
			envContext, envKubeconfig,
			false, false, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewE(tt.context, tt.kubeconfig)
			if err != nil {
				t.Fatalf("failed to create the client with context %q and kubeconfig %q", tt.context, tt.kubeconfig)
			}
			if err := c.Apply(tt.initial); (err != nil) != tt.wantInitialErr {
				t.Errorf("Client.Apply() error = %v, wantErr %v", err, tt.wantInitialErr)
				return
			}
			defer func() {
				if err := c.Delete(tt.initial); err != nil {
					t.Errorf("Client.Delete() error = %v", err)
				}
			}()

			if err := c.Apply(tt.modified); (err != nil) != tt.wantPatchErr {
				t.Errorf("Client.Apply() error = %v, wantErr %v", err, tt.wantPatchErr)
				return
			}

			got, err := tt.getChange(c)
			if (err != nil) != tt.wantGetErr {
				t.Errorf("Patch test check failed. error = %v, wantErr %v", err, tt.wantGetErr)
				return
			}
			if got != tt.want {
				t.Errorf("Client.Apply() = %v, want %v ", got, tt.want)
			}
		})
	}
}

func TestClient_ApplyFiles_thenDelete(t *testing.T) {
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
		{"apply 1 configMap file", []string{"./testdata/apply/cm.yaml"},
			func(c *Client) (bool, error) {
				cm, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-apply-0", metav1.GetOptions{})
				if err != nil {
					return false, err
				}

				isThere := (cm.Data["key1"] == "apple")
				return isThere, nil
			}, true,
			envContext, envKubeconfig, false},
		{"apply 2 configMap files", []string{"./testdata/apply/cm.yaml", "./testdata/apply/cml.yaml"},
			func(c *Client) (bool, error) {
				cm, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-apply-0", metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				cm0, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-apply-1", metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				cm1, err := c.Clientset.CoreV1().ConfigMaps("default").Get("test-apply-2", metav1.GetOptions{})
				if err != nil {
					return false, err
				}

				isThere := (cm.Data["key1"] == "apple" && cm0.Data["key1"] == "strawberry" && cm1.Data["key2"] == "pinaple")
				return isThere, nil
			}, true,
			envContext, envKubeconfig, false},
		{"apply secret file", []string{"./testdata/apply/secret.yaml"},
			func(c *Client) (bool, error) {
				s, err := c.Clientset.CoreV1().Secrets("default").Get("test-secret-apply-0", metav1.GetOptions{})
				if err != nil {
					return false, err
				}

				isThere := (string(s.Data["password"]) == "Super5ecret0!")
				return isThere, nil
			}, true,
			envContext, envKubeconfig, false},
		{"apply secret from URL", []string{"https://raw.githubusercontent.com/johandry/klient/master/testdata/apply/secret.yaml"},
			func(c *Client) (bool, error) {
				s, err := c.Clientset.CoreV1().Secrets("default").Get("test-secret-apply-0", metav1.GetOptions{})
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

			if err := c.ApplyFiles(tt.filenames...); (err != nil) != tt.wantErr {
				t.Errorf("Client.ApplyFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			isThere, err := tt.checkIsThere(c)
			if err != nil {
				t.Errorf("Client.ApplyFiles() failed to check the applied resource. error = %v", err)
			} else if isThere != tt.wantItThere {
				t.Errorf("Client.ApplyFiles() resource applied are not there")
			}

			if err := c.DeleteFiles(tt.filenames...); err != nil {
				t.Errorf("Client.Delete() error = %v", err)
			}
		})
	}
}
