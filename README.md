# Choria Prometheus File Exporter

## Overview

This is a exporter that exports data found in files on disk to Prometheus.

It includes a utility that you can call from your cronjobs and other such tools:

Increment a counter by one:

```
$ pfe count acme_adhoc_job_runs
```

Increment it by 10:

```
$ pfe count acme_adhoc_job_matches --inc 10
```

Set a gauge to an arbitrary value:

```
$ pfe gauge acme_adhoc_job_runtime 10
```

In time we'll include utilities to assist with timing events like `pfe observe acme_adhoc_job_runtime /path/to/script`.

A companion server runs that uses inotify to detect change in the data files and expose the data.

## Installation

RPMs are hosted in the Choria yum repository for el6 and 7 64bit systems:

```ini
[choria_release]
name=choria_release
baseurl=https://packagecloud.io/choria/release/el/$releasever/$basearch
repo_gpgcheck=1
gpgcheck=0
enabled=1
gpgkey=https://packagecloud.io/choria/release/gpgkey
sslverify=1
sslcacert=/etc/pki/tls/certs/ca-bundle.crt
metadata_expire=300
```

Simply installing the RPM and arranging for `prometheus-file-exporter` service to run is enough, you configure the port to listen on in `/etc/sysconfig/prometheus-file-exporter` by setting `PORT="8080"`.

## Thanks

<img src="https://packagecloud.io/images/packagecloud-badge.png" width="158">