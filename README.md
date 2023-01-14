[![](https://goreportcard.com/badge/linuxsuren/argo-workflow-atomic-plugin)](https://goreportcard.com/report/linuxsuren/argo-workflow-atomic-plugin)
[![](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/linuxsuren/argo-workflow-atomic-plugin)
[![Contributors](https://img.shields.io/github/contributors/linuxsuren/argo-workflow-atomic-plugin.svg)](https://github.com/linuxsuren/argo-workflow-atomic-plugin/graphs/contributors)
[![GitHub release](https://img.shields.io/github/release/linuxsuren/argo-workflow-atomic-plugin.svg?label=release)](https://github.com/linuxsuren/argo-workflow-atomic-plugin/releases/latest)
![GitHub All Releases](https://img.shields.io/github/downloads/linuxsuren/argo-workflow-atomic-plugin/total)

This plugin could reduce unnecessary Argo workflows. For example, there are mutiple commits against a pull request in a short time.
In most cases, only the last time of the workflow running is necessary. This plugin will stop all the workflows which have the same 
parameters and come from same WorkflowTemplate.

## Install

```shell
cat <<EOF | kubectl apply -f -
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: argo-atomic-plugin-executor-plugin
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: argo-plugin-atomic-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: argo-server-cluster-role
subjects:
- kind: ServiceAccount
  name: argo-atomic-plugin-executor-plugin
  namespace: default
- kind: ServiceAccount
  name: argo
  namespace: argo
---
apiVersion: v1
data:
  sidecar.automountServiceAccountToken: "true"
  sidecar.container: |
    image: ghcr.io/linuxsuren/argo-workflow-atomic-plugin:master
    command:
    - argo-wf-atomic
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
EOF
```

## Try it
First, create a WorkflowTemplate:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: argoproj.io/v1alpha1
kind: WorkflowTemplate
metadata:
  name: plugin-atomic
  namespace: default
spec:
  entrypoint: main
  templates:
  - name: main
    dag:
      tasks:
        - name: sleep
          template: sleep
        - name: atomic
          template: atomic
  - script:
      image: ghcr.io/linuxsuren/hd:v0.0.70
      command: [sh]
      source: sleep 90
    name: sleep
  - name: atomic
    plugin:
      argo-atomic-plugin: {}
EOF
```

then, trigger it from UI or the following command:
```shell
cat <<EOF | kubectl create -f -
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: plugin-atomic
  namespace: default
  labels:
    workflows.argoproj.io/workflow-template: plugin-atomic
spec:
  workflowTemplateRef:
    name: plugin-atomic
EOF
```
