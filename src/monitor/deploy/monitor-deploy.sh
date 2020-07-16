kubectl apply -f crds/k8splugin_v1alpha1_resourcebundlestate_crd.yaml
kubectl apply -f role.yaml
kubectl apply -f cluster_role.yaml
kubectl apply -f role_binding.yaml
kubectl apply -f clusterrole_binding.yaml
kubectl apply -f service_account.yaml
kubectl apply -f operator.yaml
