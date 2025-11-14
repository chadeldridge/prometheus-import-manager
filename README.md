# pim - prometheus-import-manager
Manage groups of targets with labels and jobs. Merge inventory sources.

/etc/pim/pim.yml
```
# Targets source file or directory to look for source files in.
# If sources is a dir pim will look from *_targets.{yml yaml json}.
sources: /etc/pim/sources

# targets_dir specifies the destination directory where the individual files will be created.
targets_dir: /etc/prometheus/file_sd

# Used to determine target file names. If the job name is node-exporter the default seetings would
# produce a targets file named node-exporter_targets.json.
# These are the default values.
#targets_file_suffix: "_targets"
#targets_file_ext: ".json"
# Might create the following target files.
#   blackbox_icmp_targets.json
#   blackbox_http_targets.json
#   blackbox_icmp_targets.json

# http_api_host specifies the ip to bind to. Default 0.0.0.0
http_api_host: 172.19.120.11
# http_api_port specifies the port to bind to. Default 9900
http_api_port: 8080
#http_tls_cert_file: ""
#http_tls_key_file: ""
http_shutdown_timeout: 5
```

Read in target configs (yml or json).
NOTE: I'm only able to combine node_exporter with blackbox in the same group because my node_exporter job configuration adds the port automatically.
/etc/pim/sources/targets.yml # Has a list with grouped targets, jobs, and labels.
```
- jobs:
    - blackbox_icmp
    - blackbox_ssh
    - node_exporter
  labels:
    environment: prod
    application: webapp
  targets:
    - atlwebapp01
    - atlwebapp02
    - atlwebapp03
- jobs:
    - blackbox_http
  labels:
    environment: prod
    application: webapp
  targets:
    - https://webapp.example.com:443
    - http://atlwebapp01.internal.com:8080
    - http://atlwebapp02.internal.com:8080
    - http://atlwebapp03.internal.com:8080
```

The above configs would produce the following files. The first 3 files will all contain the same content but with the job labelm changed.
/etc/prometheus/file_sd/blackbox_icmp_targets.json
```
[
  {
    "labels": {
      "job": "blackbox_icmp"
      "environment": "prod"
      "application": "webapp"
    },
    "targets": [
      "atlwebapp01"
      "atlwebapp02"
      "atlwebapp03"
    ]
  }
]
```
/etc/prometheus/file_sd/blackbox_ssh_targets.json
```
[
  {
    "labels": {
      "job": "blackbox_ssh"
      "environment": "prod"
      "application": "webapp"
    },
    "targets": [
      "atlwebapp01"
      "atlwebapp02"
      "atlwebapp03"
    ]
  }
]
...
```
/etc/prometheus/file_sd/node_exporter_targets.json
```
[
  {
    "labels": {
      "job": "node_exporter"
      "environment": "prod"
      "application": "webapp"
    },
    "targets": [
      "atlwebapp01"
      "atlwebapp02"
      "atlwebapp03"
    ]
  }
]
...
```
/etc/prometheus/file_sd/blackbox_http_targets.json
```
[
  {
    "labels": {
      "job": "blackbox_http"
      "environment": "prod"
      "application": "webapp"
    },
    "targets": [
      "https://webapp.example.com:443"
      "http://atlwebapp01.internal.com:8080"
      "http://atlwebapp02.internal.com:8080"
      "http://atlwebapp03.internal.com:8080"
    ]
  }
]
```

Your jobs would then point to the correct file_sd files.
file_sd:
  files:
    - blackbox_icmp_targets.yml
