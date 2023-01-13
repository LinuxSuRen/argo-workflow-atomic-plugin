[![](https://goreportcard.com/badge/linuxsuren/argo-workflow-atomic-plugin)](https://goreportcard.com/report/linuxsuren/argo-workflow-atomic-plugin)
[![](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/linuxsuren/argo-workflow-atomic-plugin)
[![Contributors](https://img.shields.io/github/contributors/linuxsuren/argo-workflow-atomic-plugin.svg)](https://github.com/linuxsuren/argo-workflow-atomic-plugin/graphs/contributors)
[![GitHub release](https://img.shields.io/github/release/linuxsuren/argo-workflow-atomic-plugin.svg?label=release)](https://github.com/linuxsuren/argo-workflow-atomic-plugin/releases/latest)
![GitHub All Releases](https://img.shields.io/github/downloads/linuxsuren/argo-workflow-atomic-plugin/total)

## Install

```shell
cat <<EOF | kubectl apply -f -
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: argo-atomic-plugin
  namespace: default
---
apiVersion: v1
data:
  sidecar.automountServiceAccountToken: "true"
  sidecar.container: |
    image: ghcr.io/linuxsuren/argo-workflow-atomic-plugin:master
    command:
    - argo-atomic
    name: argo-atomic-plugin
    ports:
    - containerPort: 3002
    resources:
      limits:
        cpu: 500m
        memory: 128Mi
      requests:
        cpu: 250m
        memory: 64Mi
    securityContext:
      allowPrivilegeEscalation: true
      runAsNonRoot: true
      runAsUser: 65534
kind: ConfigMap
metadata:
  labels:
    workflows.argoproj.io/configmap-type: ExecutorPlugin
  name: argo-atomic-plugin
  namespace: argo
```
