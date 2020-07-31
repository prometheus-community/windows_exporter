# windows_exporter on Kubernetes

With Kubernetes now supporting Windows nodes (as of [v1.14](https://kubernetes.io/docs/setup/production-environment/windows/intro-windows-in-kubernetes/)), it is useful to run this windows_exporter as a container on Windows to export metrics for your Prometheus implementation.

Please note: 
* This is a work in progress. Still need to validate the config on the Prometheus Operator
* This implementation uses port 9100. Adjust the Dockerfile as needed if port changes are desired. 
* Validate the base image against your Windows implementation. When these images do not match, the container may not run as expected.
* The windows_exporter flags and settings can be customized in the DaemonSet yaml file. 

## Container Image

Follow these steps to create a Docker image with your selected release of the windows_exporter. 

> Note. Ideally there would be a central public repository with pre-built standard images for this. 

1. Download the selected release here. https://github.com/prometheus-community/windows_exporter/releases. This Dockerfile uses the EXE, so select the appropriate one for your platform. 

2. Create your container image (this step requires a Windows machine)

    ```Dockerfile
    FROM mcr.microsoft.com/dotnet/framework/aspnet:4.8-windowsservercore-ltsc2019

    COPY windows_exporter-0.13.0-amd64.exe C:

    ENTRYPOINT [ "c:\\windows_exporter-0.13.0-amd64.exe" ]

    EXPOSE 9100
    ```

    ```bash
    docker build -t your_repo/windows_exporter:latest .
    ```

3. Push to a container registry. Follow steps to make this available in Docker Hub or your private registry.

## Kubernetes Resources

1. DaemonSet

Deploying this DaemonSet will ensure that the exporter is running on all Windows nodes in your cluster. It is important to ensure that the ```nodeSelector``` label in your Windows cluster matches the yaml below. 

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: win-node-exporter
  name: win-node-exporter
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: win-node-exporter
  template:
    metadata:
      labels:
        app: win-node-exporter
    spec:
      containers:
      - args: 
        - --collectors.enabled=os,iis,container
        - --telemetry.addr=127.0.0.1:9100
        name: win-node-exporter
        image: chzbrgr71/windows_exporter:v1.1
        ports:
        - containerPort: 9100
          hostPort: 9100
          name: https            
      nodeSelector:
        kubernetes.io/os: windows
```

Deploy: 

```bash
kubectl apply -f windows-exporter-daemonset.yaml
```

2. Service

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    k8s-app: win-node-exporter
  name: win-node-exporter
  namespace: monitoring
spec:
  type: ClusterIP
  ports:
  - name: https
    port: 9100
    protocol: TCP
    targetPort: https
  selector:
    app: win-node-exporter
```

Deploy: 

```bash
kubectl apply -f windows-exporter-service.yaml
```

3. ServiceMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    k8s-app: win-node-exporter
  name: win-node-exporter
  namespace: monitoring
spec:
  endpoints:
  - port: https
    interval: 15s
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
  jobLabel: k8s-app
  selector:
    matchLabels:
      k8s-app: win-node-exporter
```

Deploy: 

```bash
kubectl apply -f windows-exporter-servicemonitor.yaml
```