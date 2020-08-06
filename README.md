
# KubeAV: AntiVirus automation on Kubernetes

KubeAV is a Kubernetes operator that automates malware detection on Kubernetes. This is potentially useful when you're managing (and serving) untrusted data in your Kubernetes volumes.

<hr>

:warning: **COMPATIBILITY NOTICE**: This project is a prototypical implementation that is under heavy development and not considered stable. Breaking changes may occur at any time and without notice.

<hr>

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Installation](#installation)
- [CR overview](#cr-overview)
- [Usage](#usage)
  - [Starting an AV scan on demand](#starting-an-av-scan-on-demand)
  - [Scheduling an AV scan for periodic execution](#scheduling-an-av-scan-for-periodic-execution)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation

## CR overview

This operator introduces two new custom resources:

- `virusscans.av.mittwald.systems`
- `scheduledvirusscans.av.mittwald.systems`

## Usage

### Starting an AV scan on demand

An on-demand scan is modelled using the `VirusScan` custom resource (API group `av.mittwald.systems/v1beta1`). In the `.spec` of a virus scan you can specify which files to scan and which engine to use (currently, only ClamAV is supported):

```yaml
apiVersion: av.mittwald.systems/v1beta1
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

```yaml
apiVersion: av.mittwald.systems/v1beta1
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
apiVersion: av.mittwald.systems/v1beta1
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