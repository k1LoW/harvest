# Changelog

## [v0.17.3](https://github.com/k1LoW/harvest/compare/v0.17.2...v0.17.3) (2020-07-23)

* Update completion [#64](https://github.com/k1LoW/harvest/pull/64) ([k1LoW](https://github.com/k1LoW))

## [v0.17.2](https://github.com/k1LoW/harvest/compare/v0.17.1...v0.17.2) (2020-07-07)

* Fix log collect command for syslog (2 space like `Jun  6 19`) [#63](https://github.com/k1LoW/harvest/pull/63) ([k1LoW](https://github.com/k1LoW))

## [v0.17.1](https://github.com/k1LoW/harvest/compare/v0.17.0...v0.17.1) (2020-03-05)

* Fix zsh completion [#62](https://github.com/k1LoW/harvest/pull/62) ([k1LoW](https://github.com/k1LoW))

## [v0.17.0](https://github.com/k1LoW/harvest/compare/v0.16.2...v0.17.0) (2020-01-11)

* Add `hrv completion` [#61](https://github.com/k1LoW/harvest/pull/61) ([k1LoW](https://github.com/k1LoW))

## [v0.16.2](https://github.com/k1LoW/harvest/compare/v0.16.1...v0.16.2) (2019-11-27)

* Fix `hrv count` timestamp ordering [#60](https://github.com/k1LoW/harvest/pull/60) ([k1LoW](https://github.com/k1LoW))
* Use GitHub Actions [#59](https://github.com/k1LoW/harvest/pull/59) ([k1LoW](https://github.com/k1LoW))

## [v0.16.1](https://github.com/k1LoW/harvest/compare/v0.16.0...v0.16.1) (2019-10-18)

* Fix panic: runtime error: invalid memory address or nil pointer dereference [#58](https://github.com/k1LoW/harvest/pull/58) ([k1LoW](https://github.com/k1LoW))

## [v0.16.0](https://github.com/k1LoW/harvest/compare/v0.15.5...v0.16.0) (2019-10-18)

* Add `hrv info [DB_FILE]` [#57](https://github.com/k1LoW/harvest/pull/57) ([k1LoW](https://github.com/k1LoW))
* Fix timestamp grouping [#56](https://github.com/k1LoW/harvest/pull/56) ([k1LoW](https://github.com/k1LoW))
* Add table `metas` for saving `hrv fetch` info [#55](https://github.com/k1LoW/harvest/pull/55) ([k1LoW](https://github.com/k1LoW))

## [v0.15.5](https://github.com/k1LoW/harvest/compare/v0.15.4...v0.15.5) (2019-10-16)

* k8s client use --start-time and --end-time [#54](https://github.com/k1LoW/harvest/pull/54) ([k1LoW](https://github.com/k1LoW))
* Refactor RegexpParser [#53](https://github.com/k1LoW/harvest/pull/53) ([k1LoW](https://github.com/k1LoW))

## [v0.15.4](https://github.com/k1LoW/harvest/compare/v0.15.3...v0.15.4) (2019-10-15)

* Fix k8s timestamp parse [#52](https://github.com/k1LoW/harvest/pull/52) ([k1LoW](https://github.com/k1LoW))

## [v0.15.3](https://github.com/k1LoW/harvest/compare/v0.15.2...v0.15.3) (2019-10-10)

* SSH Client should fetch last line when SSH session closed. [#51](https://github.com/k1LoW/harvest/pull/51) ([k1LoW](https://github.com/k1LoW))
* Support parse 'unixtime' [#50](https://github.com/k1LoW/harvest/pull/50) ([k1LoW](https://github.com/k1LoW))

## [v0.15.2](https://github.com/k1LoW/harvest/compare/v0.15.1...v0.15.2) (2019-09-28)

* Fix: grep error Binary file (standard input) matches [#49](https://github.com/k1LoW/harvest/pull/49) ([k1LoW](https://github.com/k1LoW))

## [v0.15.1](https://github.com/k1LoW/harvest/compare/v0.15.0...v0.15.1) (2019-09-26)

* Fix configtest [#48](https://github.com/k1LoW/harvest/pull/48) ([k1LoW](https://github.com/k1LoW))

## [v0.15.0](https://github.com/k1LoW/harvest/compare/v0.14.2...v0.15.0) (2019-09-26)

* Add `hrv count` [#46](https://github.com/k1LoW/harvest/pull/46) ([k1LoW](https://github.com/k1LoW))

## [v0.14.2](https://github.com/k1LoW/harvest/compare/v0.14.1...v0.14.2) (2019-09-25)

* Fix fetch error handling [#47](https://github.com/k1LoW/harvest/pull/47) ([k1LoW](https://github.com/k1LoW))

## [v0.14.1](https://github.com/k1LoW/harvest/compare/v0.14.0...v0.14.1) (2019-09-21)

* Fix chan close timing (panic: send on closed channel) [#45](https://github.com/k1LoW/harvest/pull/45) ([k1LoW](https://github.com/k1LoW))

## [v0.14.0](https://github.com/k1LoW/harvest/compare/v0.13.1...v0.14.0) (2019-09-18)

* Do faster `hrv fetch` via ssh/file using grep timestamp [#44](https://github.com/k1LoW/harvest/pull/44) ([k1LoW](https://github.com/k1LoW))

## [v0.13.1](https://github.com/k1LoW/harvest/compare/v0.13.0...v0.13.1) (2019-09-17)

* Fix endtime (time.Time) is nil when no --end-time [#43](https://github.com/k1LoW/harvest/pull/43) ([k1LoW](https://github.com/k1LoW))

## [v0.13.0](https://github.com/k1LoW/harvest/compare/v0.12.0...v0.13.0) (2019-09-15)

* Add araddon/dataparse to parse `--*-time` option [#42](https://github.com/k1LoW/harvest/pull/42) ([k1LoW](https://github.com/k1LoW))
* Add `--duration` option [#41](https://github.com/k1LoW/harvest/pull/41) ([k1LoW](https://github.com/k1LoW))
* Fix `hrv cat` ts condition [#40](https://github.com/k1LoW/harvest/pull/40) ([k1LoW](https://github.com/k1LoW))
* Separate timestamp columns for grouping [#39](https://github.com/k1LoW/harvest/pull/39) ([k1LoW](https://github.com/k1LoW))
* Increase maxScanTokenSize [#38](https://github.com/k1LoW/harvest/pull/38) ([k1LoW](https://github.com/k1LoW))
* Refactor parser.Log struct [#37](https://github.com/k1LoW/harvest/pull/37) ([k1LoW](https://github.com/k1LoW))
* Refactor log pipeline [#36](https://github.com/k1LoW/harvest/pull/36) ([k1LoW](https://github.com/k1LoW))
* io.EOF is successful completion [#35](https://github.com/k1LoW/harvest/pull/35) ([k1LoW](https://github.com/k1LoW))

## [v0.12.0](https://github.com/k1LoW/harvest/compare/v0.11.0...v0.12.0) (2019-08-19)

* Use STDERR instead of STDOUT [#34](https://github.com/k1LoW/harvest/pull/34) ([k1LoW](https://github.com/k1LoW))
* Add gosec [#33](https://github.com/k1LoW/harvest/pull/33) ([k1LoW](https://github.com/k1LoW))
* Add --verbose option [#32](https://github.com/k1LoW/harvest/pull/32) ([k1LoW](https://github.com/k1LoW))

## [v0.11.0](https://github.com/k1LoW/harvest/compare/v0.10.0...v0.11.0) (2019-06-24)

* Add `hrv generate-k8s-config` [#31](https://github.com/k1LoW/harvest/pull/31) ([k1LoW](https://github.com/k1LoW))

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
