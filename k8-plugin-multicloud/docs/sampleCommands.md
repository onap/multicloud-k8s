# Sample Commands:

* POST
    URL:`localhost:8081/v1/vnf_instances`
    Request Body:

    ```
    {
	    "cloud_region_id": "region1",
	    "csar_id": "uuid",
        "namespace": "test",
	    "oof_parameters": [{
		    "key1": "value1",
		    "key2": "value2",
		    "key3": {}
	    }],
	    "network_parameters": {
		    "oam_ip_address": {
			    "connection_point": "string",
			    "ip_address": "string",
			    "workload_name": "string"
		    }
	    }
    }
    ```

    Expected Response:
    ```
    {
        "response": "Created Deployment:nginx-deployment"
    }
    ```

    The above POST request will download the following YAML file and run it on the Kubernetes cluster.

    ```
    apiVersion: apps/v1
    kind: Deployment
    metadata:
    name: nginx-deployment
    labels:
        app: nginx
    spec:
    replicas: 3
    selector:
        matchLabels:
        app: nginx
    template:
        metadata:
        labels:
            app: nginx
        spec:
        containers:
        - name: nginx
            image: nginx:1.7.9
            ports:
            - containerPort: 80
    ```
* GET
    URL: `localhost:8081/v1/vnf_instances`
