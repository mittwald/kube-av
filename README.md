
# KubeAV: AntiVirus automation on Kubernetes

KubeAV is a Kubernetes operator that automates malware detection on Kubernetes. This is potentially useful when you're managing (and serving) untrusted data in your Kubernetes volumes.

<hr>

:warning: **COMPATIBILITY NOTICE**: This project is a prototypical implementation that is under heavy development and not considered stable. Breaking changes may occur at any time and without notice.

<hr>

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Installation](#installation)
- [Architecture](#architecture)
- [Usage](#usage)
  - [Starting an AV scan on demand](#starting-an-av-scan-on-demand)
  - [Scheduling an AV scan for periodic execution](#scheduling-an-av-scan-for-periodic-execution)
- [Future features](#future-features)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation

Install this operator using Helm:

```
$ helm repo add mittwald https://helm.mittwald.de
$ helm repo update
$ kubectl create namespace kubeav-system
$ helm install kubeav mittwald/kube-av --namespace kubeav-system
```

## Architecture

This operator consists of several components:

- The _KubeAV operator_ runs the main controller loop. It watches for `VirusScan` and `ScheduledVirusScan` resources created by users (or itself).
- The _KubeAV updater_ is a `DaemonSet` that is created by the manager to run on every node. It maintains a local copy of the ClamAV database on each node in your cluster.
- The _KubeAV agent_ is run in `Job` resources that are managed by creating a `VirusScan` custom resource. The agent contains the actual virus scanner which uses the signature database which is maintained by the updater.

```
                            ┌────────────────┐
              creates       │ KubeAV updater │
           ┌───────────────▶│   (DaemonSet)  │
           │                └────────────────┘
┌──────────┴──────┐
│ KubeAV operator │
└──────────┬──────┘
           │  creates       ┌───────────────────┐                           ┌──────────────┐
  ┌────────────────────────▶|     VirusScan     ├──────────────────────────▶│ KubeAV agent │
  │        ├───────────────▶| (Custom Resource) │  creates (via operator)   │    (Job)     │
  │        │  watches       └───────────────────┘                           └──────────────┘
  │        │                          ▲
  │        │                          │ creates (via operator)
  │        │                          │
  │        │  creates       ┌─────────┴──────────┐
  ├────────────────────────▶│ ScheduledVirusScan │
  │        └───────────────▶│ (Custom Resource)  │
  │           watches       └────────────────────┘

  O
 /|\ User
 / \
```

## Usage

### Starting an AV scan on demand

An on-demand scan is modelled using the `VirusScan` custom resource (API group `av.mittwald.de/v1beta1`). In the `.spec` of a virus scan you can specify which files to scan and which engine to use (currently, only ClamAV is supported):

```yaml
apiVersion: av.mittwald.de/v1beta1
kind: VirusScan
metadata:
  name: example-virusscan
spec:
  # supported values: ["ClamAV"]
  engine: ClamAV

  # list of volumes to scan
  targets:

    # "volume" may be any kind of VolumeSource that you'd also use in
    # a PodSpec.
    - volume:
        persistentVolumeClaim:
            path: my-pvc
        subPath: path/to/subdir
```

A `VirusScan` resource will be mapped to a `Job` (of the `batch/v1` API group), which will in turn result in a Pod that runs the configured AV engine and that has all the specified volumes mounted.

The results of the AV scan will be written back into the `.status` property of the `VirusScan` resource:

```console
$ kubectl get virusscans
NAME                SUMMARY                        SCHEDULED   COMPLETED   AGE
example-virusscan   Completed (1 infected files)   44s         11s         44s
```

The `.status.scanResults` property in the CR lists the individual files found by the scanner:

```yaml
apiVersion: av.mittwald.de/v1beta1
kind: VirusScan
metadata:
  name: example-virusscan
spec: # ...
status:
  conditions:
    Completed:
      type: Completed
      status: "True"
    Positive:
      type: Positive
      status: "True"
  scanResults:
  # filePath:
  #   path to the infected file
  # matchingSignature:
  #   name of the detected signature as reported by the AV engine.
  - filePath: /scan/scan-target-0/infected-file
    matchingSignature: Eicar-Signature
```

### Scheduling an AV scan for periodic execution

Periodic scanning can be configured using the `ScheduledVirusScan` resource.

```yaml
apiVersion: av.mittwald.de/v1beta1
kind: ScheduledVirusScan
metadata:
  name: example-scheduledvirusscan
spec:
  # this is a standard cron schedule
  schedule: "0 */3 * * *"

  # how many "VirusScan" resources that were created from this
  # schedule should be kept.
  historySize: 3

  # template for a "VirusScan" resource
  template:
    spec:
      engine: ClamAV
      targets:
        - volume:
            hostPath:
              path: /
          subPath: root/virus
```

For these resources, KubeAV will create new `VirusScan` resources from the configured template at the specified interval.

## Future features

- [ ] Alerting (maybe by adding a metric for counting infected files? Or by directly integrating something like the Prometheus alert manager)
- [ ] [On-Access Scanning](https://www.clamav.net/documents/on-access-scanning) (Todo: determine feasibility in containerized environment)
