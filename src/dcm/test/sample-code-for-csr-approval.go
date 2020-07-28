package main

import (
        "context"
        "flag"
        "fmt"
        "os"
        "path/filepath"
        "time"

        "k8s.io/apimachinery/pkg/api/errors"
        certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/client-go/kubernetes"
        "k8s.io/client-go/tools/clientcmd"
)

func main() {
        var kubeconfig *string
		kubeconfig = flag.String("kubeconfig", filepath.Join("/root/", ".kube", "config"), "(optional) absolute path to the kubeconfig file")
        flag.Parse()

        config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
        if err != nil {
                panic(err.Error())
        }

        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                panic(err.Error())
        }

		// 1st. need a copy of existing CSR
        csr, err := clientset.CertificatesV1beta1().CertificateSigningRequests().Get(context.TODO(), "lc1-user-csr", metav1.GetOptions{})

        csr.Status.Conditions = append(csr.Status.Conditions, certificatesv1beta1.CertificateSigningRequestCondition{
                Type: certificatesv1beta1.CertificateApproved,
                Reason: "ApprovedForDCM",
                Message: "Approved for DCM",
                LastUpdateTime: metav1.Now(),
        })

		// 2nd. approve CSR using UpdateApproval()
        newcsr, err := clientset.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(context.TODO(), csr, metav1.UpdateOptions{})
}
