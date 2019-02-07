# Harvest

> Portable log aggregation tool for middle scale system operation/observation.

Harvest provides the `hrv` command with the following features.

- Fetch remote/local logs to SQLite database via SSH/exec ( `hrv fetch` )
- Output logs from SQLite database ( `hrv cat` )

## Usage

### 1. Set log URLs (and log type) in config.yml

``` yaml
---
logs:
  -
    url: 'ssh://webproxy.example.com/var/log/syslog*'
    description: webproxy syslog
    type: syslog
    tags:
      - webproxy
      - syslog
  -
    url: 'ssh://webproxy.example.com/var/log/nginx/access_log*'
    description: webproxy NGINX access log
    type: combinedLog
    tags:
      - webproxy
      - nginx
  -
    url: 'ssh://app-1.example.com/var/log/ltsv.log*'
    description: app-1 log
    type: regexp
    regexp: 'time:([^\t]+)'
    timeFormat: 'Jan 02 15:04:05'
    timeZone: '+0900'
    tags:
      - app
  -
    url: 'ssh://app-2.example.com/var/log/ltsv.log*'
    description: app-2 log
    type: regexp
    regexp: 'time:([^\t]+)'
    timeFormat: 'Jan 02 15:04:05'
    timeZone: '+0900'
    tags:
      - app
  -
    url: 'ssh://db.example.com/var/log/tcpdp/eth0/dump*'
    description: db dump log
    type: regexp
    regexp: '"ts":"([^"]+)"'
    timeFormat: '2006-01-02T15:04:05.999-0700'
    tags:
      - db
      - query
  -
    url: 'ssh://db.example.com/var/log/postgresql/postgresql*'
    description: PostgreSQL log
    type: regexp
    regexp: '^\[?(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [^ ]{3})'
    multiLine: true
    timeFormat: '2006-01-02 15:04:05 MST'
    tags:
      - db
      - postgresql
  -
    url: 'file:///path/to/httpd/access.log'
    description: local Apache access log
    type: combinedLog
    tags:
      - httpd
```

### 2. Fetch target logs via SSH/exec ( `hrv fecth` )

``` console
$ hrv fetch -c config.yml
```

### 3. Output logs ( `hrv cat` )

``` console
$ hrv cat harvest-20181215T2338+900.db --with-timestamp --with-host --with-path | less -R
```

## Requirements

- awk
- date
- find
- grep
- ls
- sudo
- xargs
- zcat

## TODO

- `fetch-check` command

## References

- [Hayabusa](https://github.com/hirolovesbeer/hayabusa): A Simple and Fast Full-Text Search Engine for Massive System Log Data
