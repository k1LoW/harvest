# Harvest [![Build Status](https://travis-ci.org/k1LoW/harvest.svg?branch=master)](https://travis-ci.org/k1LoW/harvest) [![GitHub release](https://img.shields.io/github/release/k1LoW/harvest.svg)](https://github.com/k1LoW/harvest/releases) [![Go Report Card](https://goreportcard.com/badge/github.com/k1LoW/harvest)](https://goreportcard.com/report/github.com/k1LoW/harvest)

> Portable log aggregation tool for middle-scale system operation/troubleshooting.

![screencast](doc/screencast.svg)

Harvest provides the `hrv` command with the following features.

- Agentless.
- Portable.
- Only 1 config file.
- Fetch various remote/local log data via SSH/exec/Kubernetes API. ( `hrv fetch` )
- Output all fetched logs in the order of timestamp. ( `hrv cat` )
- Stream various remote/local logs via SSH/exec/Kubernetes API. ( `hrv stream` )
- Copy remote/local raw logs via SSH/exec. ( `hrv cp` )

## Quick Start ( for Kubernetes )

``` console
$ hrv generate-k8s-config > cluster.yml
$ hrv stream -c cluster.yml --tag='kube_apiserver or coredns' --with-path --with-timestamp
```

## Usage

### :beetle: Fetch and output remote/local log data

#### 1. Set log sources (and log type) in config.yml

``` yaml
---
targetSets:
  -
    description: webproxy syslog
    type: syslog
    sources:
      - 'ssh://webproxy.example.com/var/log/syslog*'
    tags:
      - webproxy
      - syslog
  -
    description: webproxy NGINX access log
    type: combinedLog
    sources:
      - 'ssh://webproxy.example.com/var/log/nginx/access_log*'
    tags:
      - webproxy
      - nginx
  -
    description: app log
    type: regexp
    regexp: 'time:([^\t]+)'
    timeFormat: 'Jan 02 15:04:05'
    timeZone: '+0900'
    sources:
      - 'ssh://app-1.example.com/var/log/ltsv.log*'
      - 'ssh://app-2.example.com/var/log/ltsv.log*'
      - 'ssh://app-3.example.com/var/log/ltsv.log*'
    tags:
      - app
  -
    description: db dump log
    type: regexp
    regexp: '"ts":"([^"]+)"'
    timeFormat: '2006-01-02T15:04:05.999-0700'
    sources:
      - 'ssh://db.example.com/var/log/tcpdp/eth0/dump*'
    tags:
      - db
      - query
  -
    description: PostgreSQL log
    type: regexp
    regexp: '^\[?(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} \w{3})'
    timeFormat: '2006-01-02 15:04:05 MST'
    multiLine: true
    sources:
      - 'ssh://db.example.com/var/log/postgresql/postgresql*'
    tags:
      - db
      - postgresql
  -
    description: local Apache access log
    type: combinedLog
    sources:
      - 'file:///path/to/httpd/access.log'
    tags:
      - httpd
-
    description: api on Kubernetes
    type: k8s
    sources:
      - 'k8s://context-name/namespace/pod-name*'
    tags:
      - api
      - k8s
```

You can use `hrv configtest` for config test.

``` console
$ hrv configtest -c config.yml
```

#### 2. Fetch target log data via SSH/exec/Kubernetes API ( `hrv fecth` )

``` console
$ hrv fetch -c config.yml --tag=webproxy,db
```

#### 3. Output log data ( `hrv cat` )

``` console
$ hrv cat harvest-20181215T2338+900.db --with-timestamp --with-host --with-path | less -R
```

#### 4. Count log data ( `hrv count` )

``` console
$ hrv count harvest-20191015T2338+900.db -g minute -g webproxy -b db
ts      webproxy db
2019-09-24 08:01:00     9618    5910
2019-09-24 08:02:00     9767    5672
2019-09-24 08:03:00     10815   7394
2019-09-24 08:04:00     11782   7109
2019-09-24 08:05:00     9896    6346
[...]
2019-09-24 08:24:00     11619   5646
2019-09-24 08:25:00     10541   6097
2019-09-24 08:26:00     11336   5264
2019-09-24 08:27:00     1102    5261
2019-09-24 08:28:00     1318    6660
2019-09-24 08:29:00     10362   5663
2019-09-24 08:30:00     11136   5373
2019-09-24 08:31:00     1748    1340
```

### :beetle: Stream remote/local logs

#### 1. [Set config.yml](#1-set-log-sources-and-log-type-in-configyml)

#### 2. Stream target logs via SSH/exec/Kubernetes API ( `hrv stream` )

``` console
$ hrv stream -c config.yml --with-timestamp --with-host --with-path --with-tag
```

### :beetle: Copy remote/local raw logs

#### 1. [Set config.yml](#1-set-log-sources-and-log-type-in-configyml)

#### 2. Copy remote/local raw logs to local directory via SSH/exec ( `hrv cp` )

``` console
$ hrv cp -c config.yml
```

### --tag filter operators

The following operators can be used to filter targets

`not`, `and`, `or`, `!`, `&&`, `||`

``` console
$ hrv stream -c config.yml --tag='webproxy or db' --with-timestamp --with-host --with-path
```

#### `,` is converted to ` or `

``` console
$ hrv stream -c config.yml --tag='webproxy,db'
```

is converted to

``` console
$ hrv stream -c config.yml --tag='webproxy or db'
```

### --source filter

filter targets using source regexp

``` console
$ hrv fetch -c config.yml --source='app-[0-9].example'
```

## Architecture

### `hrv fetch` and `hrv cat`

![img](doc/fetch.png)

### `hrv stream`

![img](doc/stream.png)

## Installation

```console
$ brew install k1LoW/tap/harvest
```

or

```console
$ go get github.com/k1LoW/harvest/cmd/hrv
```

## What is "middle-scale system"?

- < 50 instances
- < 1 million logs per `hrv fetch`

### What if you are operating a large-scale/super-large-scale/hyper-large-scale system?

Let's consider agent-base log collector/platform, service mesh and distributed tracing platform!

## Internal

- [harvest-*.db database schema](doc/schema)

## Requirements

- UNIX commands
  - date
  - find
  - grep
  - head
  - ls
  - tail
  - xargs
  - zcat
- sudo
- SQLite

## WANT

- tag DAG
- Viewer / Visualizer

## References

- [Hayabusa](https://github.com/hirolovesbeer/hayabusa): A Simple and Fast Full-Text Search Engine for Massive System Log Data
    - Make simple with a combination of commands.
    - Full-Text Search Engine using SQLite FTS.
- [stern](https://github.com/wercker/stern): ⎈ Multi pod and container log tailing for Kubernetes
    - Multiple Kubernetes log streaming architecture.
