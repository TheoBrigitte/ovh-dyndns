# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/)
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Add support for IPv6 lookup
- Add support for multiple records in config file
- Add handling for consensus.UseIPProtocol error
- Add handling for viper.BindPFlags error
- Add tests for hostname and public IP
- Add README.md and sample configuration
- Add Makefile
- Add Go modules

### Changed

- Improve ip.GetPublic with more robust providers
- Create record if it does not exists
- Update DNS record update loop to continue and print errors rather than stop with fatal on error
- Replace --version implementation with prometheus/common/verion package
- Refresh DNS zone after each update/create
- Only update DNS record when target do not match
- Validate IP address format
- Update github.com/ovh/go-ovh to v1.3.0

### Fixed

- Fix ipAddress variable name typo
- Fix linter errors

[Unreleased]: https://github.com/TheoBrigitte/ovh-dyndns/tree/main
