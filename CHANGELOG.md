# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Fix linting issues.

## [4.12.0] - 2025-03-21

### Added

- Add BackupDestination label to support multiple operator instances

## [4.11.0] - 2025-03-20

### Added

- Add toggle to skip CRD installation via `crds.install: false` when deploying multiple instances of the operator.

## [4.10.1] - 2025-03-18

### Fixed

- Ensure the default `clustersToExcludeRegex` doesn't match any clusters

## [4.10.0] - 2025-02-27

### Added

- Add support for excluding clusters by defining a regex.

## [4.9.0] - 2025-01-30

### Added

- Support for other s3 compatible storages

## [4.8.2] - 2024-07-22

### Changed

- Go: Dependency updates.

## [4.8.1] - 2024-06-26

### Changed

- Repository: Some chores. ([#654](https://github.com/giantswarm/etcd-backup-operator/pull/654))
  - Go: Update dependencies.
  - Repository: Rework `create-cr.sh`.
  - Repository: Rework `Dockerfile`.
  - Chart: Rework CronJob.

## [4.8.0] - 2024-06-26

### Added

- Support skipping backup of Vintage AWS clusters by adding the annotation `giantswarm.io/etcd-backup-operator-skip-backup=true` to the `AWSCluster` object. This can be used for clusters which got migrated to CAPI.

### Changed

- Remove deprecated packages and grpc DialOption.

## [4.7.0] - 2024-04-01

### Changed

- Use ServiceMonitor for monitoring.

## [4.6.0] - 2024-01-17

### Changed

- Opt-out PSP for Kubernetes v1.25.

## [4.5.0] - 2024-01-16

### Changed

- Skip reconciling etcdbackup CR's if a newer one is pending.
- Configure `gsoci.azurecr.io` as the default container image registry.

## [4.4.6] - 2023-12-11

## [4.4.5] - 2023-12-07

### Changed

- Changed ownership to team-turtles.

## [4.4.4] - 2023-12-07

### Changed

- Refactor how we get etcd endpoints in CAPI clusters.
- Packages updates.

## [4.4.3] - 2023-11-23

### Changed

- Set 50m min cpu in VPA

## [4.4.2] - 2023-11-06

- Add PolicyException to `etcd-backup-operator-scheduler`.

## [4.4.1] - 2023-10-23

### Changed

- Add helm hook annotations to PolicyExceptions CRs.
- Update `golang.org/x/net` package to `v0.13.0`.
- Update `google.golang.org/grpc` package to `v1.57.0`.

## [4.4.0] - 2023-07-13

### Added

- Added required values for pss policies.
- Added pss exceptions for volumes and ports.

### Removed

- Stop pushing to `openstack-app-collection`.

## [4.3.1] - 2023-05-03

### Changed

- Fix kubernetes version check to add toleration for `node-role.kubernetes.io/control-plane`.

## [4.3.0] - 2023-02-20

### Added

- Added the use of the runtime/default seccomp profile.
- Added option to set `etcdBackupEncryptionPassword` which enables encryption of the backup.

## [4.2.1] - 2023-01-17

### Fixed

- Correctly mark a backup task as failed when etcd client can't be initialized.
- Check cluster domain is set or fail backup early.

## [4.2.0] - 2023-01-17

### Changed

- Log error reason when the preparation for v3 backup fails.

## [4.1.0] - 2022-11-02

### Changed

- Added option to set pod's `priorityClassName`.
- Use `github.com/nats-io/nats-server` version `v2.9.3` and `golang.org/x/text` version `v0.3.8` to avoid vulnerabilities.

### Fixed

- `etcd-backup-operator` is now compatible with Kubernetes Versions >= `v1.24`

## [4.0.0] - 2022-09-20

### Added

- Added CRD to helm chart.

## [3.2.0] - 2022-07-06

- Add functionality to backup CAPI clusters.

## [3.1.0] - 2022-06-15

### Added

- Extend the operator to allow multiple schedules and select which clusters will be backed up.

## [3.0.1] - 2022-04-04

### Fixed

- Bump go module version in `go.mod`.

## [3.0.0] - 2022-03-31

- Use `giantswarm/k8smetadata` for labels.
- Update `giantswarm/apiextensions` to `v6.0.0`.
- Update k8s dependencies to `v0.22.2`.

## [2.10.1] - 2022-03-22

### Fixed

- Ignore VPA configuration if VPA is not installed.

## [2.10.0] - 2022-02-24

### Changed

- Disabled Sentry

## [2.9.1] - 2022-02-07

### Fixed

- The `revision` data coming from `etcdctl` needs an `int64` to fit.
- Fix nil pointer in collector.

## [2.9.0] - 2022-02-03

### Changed

- Allow container port to be configured.
- Switch default container port to 8050 to avoid port collisions.

### Changed

- Switch from apiextensions to apiextensions-backup for etcdbackup CRD.

## [2.8.0] - 2022-01-21

- Add possibility to backup specific clusters within an installation.

## [2.7.2] - 2022-01-17

### Fixed

- Fixed RBAC

## [2.7.1] - 2022-01-17

### Fixed

- Fix etcd certs lookup to search for secrets in all namespaces.

## [2.7.0] - 2021-11-24

### Added

- Run 'compact' and 'defrag' on each etcd instance before taking the v3 backup.

## [2.6.0] - 2021-11-18

### Changed

- Look for cluster certificates in the cluster namespace ( instead of looking only in default namespace).

## [2.5.0] - 2021-11-16

### Changed

- Smart apiVersion selection for cronjob.
- Use a clearer naming schema for backup files.

### Added

- Added `values.schema.json` file to validate helm values.

## [2.4.0] - 2021-08-05

### Changed

- Prepare helm values to configuration management.
- Update architect-orb to v3.0.0.

## [2.3.0] - 2021-07-19

### Changed

- Disabled etcd v2 backup for Azure and AWS.

## [2.2.1] - 2021-04-06

- Bump up dependencies:
  - OperatorKit `v4.3.1`.
  - k8sclient `v5.11.0`
  - apiextensions `v3.22.0`

## [2.2.0] - 2021-02-26

### Added

- Added vertical pod autoscaler to the helm chart.

## [2.1.0] - 2020-12-04

### Added

- Add sentry support.

## [2.0.1] - 2020-11-25

### Fixed

- Add support for 13.0.0 azure clusters.

## [2.0.0] - 2020-09-23

### Changed

- Updated backward incompatible Kubernetes dependencies to v1.18.5.

## [1.2.0] - 2020-09-18

### Added

- Add monitoring labels

### Changed

- Use a different secret to get ETCD data from because the previous approach wasn't working for HA masters clusters.

## [1.1.0] 2020-08-21

### Added

- Label Jobs spawned from CronJob.
- Add NetworkPolicy.

## [1.0.5] 2020-06-01

### Fixed

- Consider clusters for all providers in the prometheus exporter.

## [1.0.4] 2020-05-22

### Fixed

- Avoid exporting metrics for deleted clusters.

## [1.0.3] 2020-05-22

### Fixed

- Fixed prometheus exporter.

## [1.0.2] 2020-05-22

### Fixed

- Fix name in helm chart Secret template.

## [1.0.1] 2020-05-21

### Fixed

- Fix version in project.go.

## [1.0.0] 2020-05-19

### Added

- First release.

[Unreleased]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.12.0...HEAD
[4.12.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.11.0...v4.12.0
[4.11.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.10.1...v4.11.0
[4.10.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.10.0...v4.10.1
[4.10.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.9.0...v4.10.0
[4.9.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.8.2...v4.9.0
[4.8.2]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.8.1...v4.8.2
[4.8.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.8.0...v4.8.1
[4.8.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.7.0...v4.8.0
[4.7.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.6.0...v4.7.0
[4.6.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.5.0...v4.6.0
[4.5.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.4.6...v4.5.0
[4.4.6]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.4.5...v4.4.6
[4.4.5]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.4.4...v4.4.5
[4.4.4]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.4.3...v4.4.4
[4.4.3]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.4.2...v4.4.3
[4.4.2]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.4.1...v4.4.2
[4.4.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.4.0...v4.4.1
[4.4.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.3.1...v4.4.0
[4.3.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.3.0...v4.3.1
[4.3.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.2.1...v4.3.0
[4.2.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.2.0...v4.2.1
[4.2.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.1.0...v4.2.0
[4.1.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v4.0.0...v4.1.0
[4.0.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v3.2.0...v4.0.0
[3.2.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v3.1.0...v3.2.0
[3.1.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v3.0.1...v3.1.0
[3.0.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v3.0.0...v3.0.1
[3.0.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.10.1...v3.0.0
[2.10.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.10.0...v2.10.1
[2.10.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.9.1...v2.10.0
[2.9.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.9.0...v2.9.1
[2.9.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.8.0...v2.9.0
[2.8.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.7.2...v2.8.0
[2.7.2]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.7.1...v2.7.2
[2.7.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.7.0...v2.7.1
[2.7.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.6.0...v2.7.0
[2.6.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.5.0...v2.6.0
[2.5.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.4.0...v2.5.0
[2.4.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.3.0...v2.4.0
[2.3.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.2.1...v2.3.0
[2.2.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.2.0...v2.2.1
[2.2.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.0.1...v2.1.0
[2.0.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.2.0...v2.0.0
[1.2.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.0.5...v1.1.0
[1.0.5]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/giantswarm/etcd-backup-operator/releases/tag/v1.0.0
