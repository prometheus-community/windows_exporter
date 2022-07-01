# windows_exporter on Kubernetes

With Kubernetes supporting HostProcess containers on Windows nodes (as of [v1.22](https://kubernetes.io/blog/2021/08/16/windows-hostprocess-containers/), it is useful to run the `windows_exporter` as a container on Windows to export metrics for your Prometheus implementation.  Read the [Kubernetes HostProcess documentation](https://kubernetes.io/docs/tasks/configure-pod-container/create-hostprocess-pod/) for more information.

Requirements:

- Kubernetes 1.22+
- containerd 1.6 Beta+
- WindowsHostProcessContainers feature-gate turned on for `kube-apiserver` and `kubelet`

> IMPORTANT: This does not work unless you are specifically targeting Host Process Containers with Containerd (Docker doesn't have support).  The image will build but will **not** be able to access the host.

## Container Image

The image is multi arch image (WS 2019, WS 2022) built on Windows. To build the images:

```
DOCKER_REPO=<your repo> make push-all
```

If you don't have a version of `make` on your Windows machine, You can use WSL to build the image with Windows Containers by creating a symbolic link to the docker cli and then override the docker command in the `Makefile`: 

On Windows Powershell prompt:
```
New-Item -ItemType SymbolicLink -Path "c:\docker" -Target "C:\Program Files\Docker\Docker\resources\bin\docker.exe"
```

In WSL:
```
DOCKER_REPO=<your repo> DOCKER=/mnt/c/docker make push-all 
```

## Kubernetes Quick Start

Before beginning you need to deploy the [prometheus operator](https://github.com/prometheus-operator/prometheus-operator) to your cluster. As a quick start, you can use a project like https://github.com/prometheus-operator/kube-prometheus. The export itself doesn't have any dependency on prometheus operator and the exporter image can be used in manual configurations.

### Windows Exporter DaemonSet

This create a deployment on every node. A config map is created for to handle the configuration of the Windows exporter with [configuration file](../README.md#using-a-configuration-file).  Adjust the configuration file for the collectors you are interested in.

```bash
kubectl apply -f kubernetes/windows-exporter-daemonset.yaml
```

> Note: This example manifest deploys the latest bleeding edge image `ghcr.io/prometheus-community/windows-exporter:latest` built from the main branch.  You should update this to use a released version which you can find at https://github.com/prometheus-community/windows_exporter/releases

#### Configuring the firewall
The firewall on the node needs to be configured  to allow connections on the node: `New-NetFirewallRule -DisplayName 'windows-exporter' -Direction inbound -Profile Any -Action Allow -LocalPort 9182 -Protocol TCP` 

You could do this by adding an init container but if you remove the deployment at a later date you will need to remove the firewall rule manually. The following could be added to the `windows-exporter-daemonset.yaml`:

```
apiVersion: apps/v1
kind: DaemonSet
spec:
  template:
    spec:
      initContainers:
        - name: configure-firewall
          image: mcr.microsoft.com/windows/powershell:lts-nanoserver-1809
          command: ["powershell"]
          args: ["New-NetFirewallRule", "-DisplayName", "'windows-exporter'", "-Direction", "inbound", "-Profile", "Any", "-Action", "Allow", "-LocalPort", "9182", "-Protocol", "TCP"]
```

### Prometheus PodMonitor

Create the [Pod Monitor](https://prometheus-operator.dev/docs/operator/design/#podmonitor) to configure the scraping:

```bash
kubectl apply -f windows-exporter-podmonitor.yaml
```

### View Metrics

Open Prometheus with 

```
kubectl --namespace monitoring port-forward svc/prometheus-k8s 9091:9090
```

Navigate to prometheus UI and add a query to see node cpu (replacing with your ip address)

```
sum by (mode) (irate(windows_cpu_time_total{instance="10.1.0.5:9182"}[5m]))
```

![windows cpu total time graph in prometheus ui](https://user-images.githubusercontent.com/648372/140547130-b535c766-6479-47d3-b2d3-cd8a551647df.png)


## Configuring TLS

It is possible to configure TLS of the solution using `--web.config.file`.  Read more at https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md
