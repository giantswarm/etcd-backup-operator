# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.0.4...HEAD
[1.0.4]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/giantswarm/etcd-backup-operator/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/giantswarm/etcd-backup-operator/releases/tag/v1.0.0
