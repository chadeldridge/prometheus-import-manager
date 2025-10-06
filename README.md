# prometheus-target-manager
Manage groups of targets with labels and jobs. Merge inventory sources.

Create initial file_sd provisioner then add http exporter.

/etc/pim/pim.yml
```
# Directory to look for template files.
templates_dir: /etc/pim/templates

# File name for the target templates file. pim will look for the file in templates_dir.
target_templates_file: targets.yml

# file_sd_dir specifies the target file destination directory where the individual files will be
# created. Right now you can only specify a single destination.
file_sd_dir: /etc/prometheus/file_sd

# target_split tells pim how to name the files or http endpoints according to the labels in
# the target.yml file.
# target_split:
#   - job
#   - datacenter
#   - application
# Might create the following target files.
#   blackbox_icmp_atl_webapp.json
#   blackbox_http_atl_webapp.json
#   blackbox_icmp_jfk_mysql.json
target_split:
  - job
```

Read in target configs (yml or json).
NOTE: I'm only able to combine node_exporter with blackbox in the same group because my node_exporter job configuration adds the port automatically.
/etc/pim/templates/targets.yml # Has a list with grouped targets, jobs, and labels.
```
target_groups:
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
      "job": "blackbox_icmp"
...
```
/etc/prometheus/file_sd/node_exporter_targets.json
```
[
  {
    "labels": {
      "job": "node_exporter"
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
