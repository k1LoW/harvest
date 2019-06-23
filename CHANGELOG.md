# Changelog

## [v0.10.0](https://github.com/k1LoW/harvest/compare/v0.9.0...v0.10.0) (2019-06-23)

* Support Kubernetes logs [#30](https://github.com/k1LoW/harvest/pull/30) ([k1LoW](https://github.com/k1LoW))

## [v0.9.0](https://github.com/k1LoW/harvest/compare/v0.8.0...v0.9.0) (2019-06-18)

* [BREAKING] Update schema [#29](https://github.com/k1LoW/harvest/pull/29) ([k1LoW](https://github.com/k1LoW))

## [v0.8.0](https://github.com/k1LoW/harvest/compare/v0.7.0...v0.8.0) (2019-06-16)

* Change `--url-regexp` -> `--source` [#28](https://github.com/k1LoW/harvest/pull/28) ([k1LoW](https://github.com/k1LoW))
* Fix `hrv targets` output ( when source is `file://` ) [#27](https://github.com/k1LoW/harvest/pull/27) ([k1LoW](https://github.com/k1LoW))

## [v0.7.0](https://github.com/k1LoW/harvest/compare/v0.6.3...v0.7.0) (2019-06-16)

* Add `hrv tags` [#26](https://github.com/k1LoW/harvest/pull/26) ([k1LoW](https://github.com/k1LoW))
* [BREAKING CHANGE] Change filter logic of `--tag` option  [#25](https://github.com/k1LoW/harvest/pull/25) ([k1LoW](https://github.com/k1LoW))
* [BREAKING CHANGE] Change command [#24](https://github.com/k1LoW/harvest/pull/24) ([k1LoW](https://github.com/k1LoW))
* [BREAKING] Change config format `urls:` -> `sources:` [#23](https://github.com/k1LoW/harvest/pull/23) ([k1LoW](https://github.com/k1LoW))

## [v0.6.3](https://github.com/k1LoW/harvest/compare/v0.6.2...v0.6.3) (2019-05-17)

* Fix target filter [#22](https://github.com/k1LoW/harvest/pull/22) ([k1LoW](https://github.com/k1LoW))
* Skip configtest when target type = 'none' [#21](https://github.com/k1LoW/harvest/pull/21) ([k1LoW](https://github.com/k1LoW))

## [v0.6.2](https://github.com/k1LoW/harvest/compare/v0.6.1...v0.6.2) (2019-03-16)

* Fix goreleaser.yml for  for CGO_ENABLED=1 [#20](https://github.com/k1LoW/harvest/pull/20) ([k1LoW](https://github.com/k1LoW))

## [v0.6.1](https://github.com/k1LoW/harvest/compare/v0.6.0...v0.6.1) (2019-03-16)

* Fix `file://` log aggregation [#19](https://github.com/k1LoW/harvest/pull/19) ([k1LoW](https://github.com/k1LoW))
* Fix CGO_ENABLED [#18](https://github.com/k1LoW/harvest/pull/18) ([k1LoW](https://github.com/k1LoW))

## [v0.6.0](https://github.com/k1LoW/harvest/compare/v0.5.0...v0.6.0) (2019-03-16)

* Use goreleaser [#17](https://github.com/k1LoW/harvest/pull/17) ([k1LoW](https://github.com/k1LoW))
* Fix maxContentStash logic [#16](https://github.com/k1LoW/harvest/pull/16) ([k1LoW](https://github.com/k1LoW))
* Add parser type `none` [#15](https://github.com/k1LoW/harvest/pull/15) ([k1LoW](https://github.com/k1LoW))
* Add `--without-mark` option [#14](https://github.com/k1LoW/harvest/pull/14) ([k1LoW](https://github.com/k1LoW))

## [v0.5.0](https://github.com/k1LoW/harvest/compare/v0.4.0...v0.5.0) (2019-02-27)

* Preset default passphrase for all targets [#13](https://github.com/k1LoW/harvest/pull/13) ([k1LoW](https://github.com/k1LoW))

## [v0.4.0](https://github.com/k1LoW/harvest/compare/v0.3.0...v0.4.0) (2019-02-21)

* Add `hrv ls-logs` [#12](https://github.com/k1LoW/harvest/pull/12) ([k1LoW](https://github.com/k1LoW))
* Change config.yml format ( `logs:` -> `targetSets:` ) [#11](https://github.com/k1LoW/harvest/pull/11) ([k1LoW](https://github.com/k1LoW))
* Add `hrv cp` [#10](https://github.com/k1LoW/harvest/pull/10) ([k1LoW](https://github.com/k1LoW))
* Fix configtest targets order [#9](https://github.com/k1LoW/harvest/pull/9) ([k1LoW](https://github.com/k1LoW))
* Add `--preset-ssh-key-passphrase` option [#8](https://github.com/k1LoW/harvest/pull/8) ([k1LoW](https://github.com/k1LoW))

## [v0.3.0](https://github.com/k1LoW/harvest/compare/v0.2.3...v0.3.0) (2019-02-19)

* Add `hrv ls-targets` [#7](https://github.com/k1LoW/harvest/pull/7) ([k1LoW](https://github.com/k1LoW))

## [v0.2.3](https://github.com/k1LoW/harvest/compare/v0.2.2...v0.2.3) (2019-02-15)

* Fix build*Command ( append more `sudo` ) [#6](https://github.com/k1LoW/harvest/pull/6) ([k1LoW](https://github.com/k1LoW))

## [v0.2.2](https://github.com/k1LoW/harvest/compare/v0.2.1...v0.2.2) (2019-02-15)

* Show MultiLine when configtest error [#5](https://github.com/k1LoW/harvest/pull/5) ([k1LoW](https://github.com/k1LoW))
* Show error message when log read error [#4](https://github.com/k1LoW/harvest/pull/4) ([k1LoW](https://github.com/k1LoW))
* Fix build command [#3](https://github.com/k1LoW/harvest/pull/3) ([k1LoW](https://github.com/k1LoW))

## [v0.2.1](https://github.com/k1LoW/harvest/compare/v0.2.0...v0.2.1) (2019-02-14)

* Change directories [#2](https://github.com/k1LoW/harvest/pull/2) ([k1LoW](https://github.com/k1LoW))

## [v0.2.0](https://github.com/k1LoW/harvest/compare/51449d0b6a46...v0.2.0) (2019-02-14)

* Add `hrv configtest` [#1](https://github.com/k1LoW/harvest/pull/1) ([k1LoW](https://github.com/k1LoW))

## [v0.1.0](https://github.com/k1LoW/harvest/compare/51449d0b6a46...v0.1.0) (2019-02-13)
