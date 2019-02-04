# Harvest

> Portable log aggregation tool for middle scale system operation/observation.

Harvest provides the `hrv` command with the following features.

- Fetch remote/local logs to SQLite database via SSH/exec ( `hrv fetch` )
- Output logs from SQLite database ( `hrv cat` )

## Usage

### 1. Set log URLs (and log type) in config.yml

``` yaml
---
targets:
  -
    url: 'ssh://webproxy.example.com/var/log/syslog'
    description: webproxy syslog
    type: syslog
  -
    url: 'ssh://webproxy.example.com/var/log/nginx/access_log'
    description: webproxy NGINX access log
    type: combinedLog
  -
    url: 'ssh://app-1.example.com/var/log/ltsv.log'
    description: app-1 log
    type: regexp
    regexp: 'time:([^\t]+)'
    timeFormat: 'Jan 02 15:04:05'
    timeZone: '+0900'
  -
    url: 'ssh://app-2.example.com/var/log/ltsv.log'
    description: app-2 log
    type: regexp
    regexp: 'time:([^\t]+)'
    timeFormat: 'Jan 02 15:04:05'
    timeZone: '+0900'
  -
    url: 'ssh://db.example.com/var/log/tcpdp/eth0/dump.log'
    description: db dump log
    type: regexp
    regexp: '"ts":"([^"]+)"'
    timeFormat: '2006-01-02T15:04:05.999-0700'
  -
    url: 'file:///path/to/httpd/access.log'
    description: local Apache access log
    type: combinedLog
```

### 2. `hrv fecth`: fetch logs from targets

``` console
$ hrv fetch -c config.yml -o harvest.db
```

### 3. `hrv cat`: cat logs

``` console
$ hrv cat harvest.db --with-ts --with-host
```

## TODO

- Target filter option like `--host 'app-*'` or label/tag
- Find lotated logs
- Support multi-line log

## References

- [Hayabusa](https://github.com/hirolovesbeer/hayabusa): A Simple and Fast Full-Text Search Engine for Massive System Log Data
