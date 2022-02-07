# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

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

[Unreleased]: https://github.com/giantswarm/giantswarm/compare/v2.9.0...HEAD
[2.9.0]: https://github.com/giantswarm/giantswarm/compare/v2.8.0...v2.9.0
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
