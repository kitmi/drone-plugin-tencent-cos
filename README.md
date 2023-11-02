The plugin upload files to your tencent cos bucket.

# Usage

The following settings changes this plugin's behavior.

* command: (optional) default as upload
* bucket: storage bucket name
* region: storage region
* source: source path 
* target: target path

Below is an example `.drone.yml` that uses this plugin.

```yaml
kind: pipeline
name: default

steps:
- name: run kitmi/tencent-cos-plugin plugin
  image: kitmi/tencent-cos-plugin
  pull: if-not-exists
  settings:
    command: download
    bucket: foo
    region: bar
    source: conf/sso
    target: app/app.conf
```

# Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t kitmi/tencent-cos-plugin -f docker/Dockerfile .
```

# Testing

Execute the plugin from your current working directory:

```text
docker run --rm -e PLUGIN_PARAM1=foo -e PLUGIN_PARAM2=bar \
  -e DRONE_COMMIT_SHA=8f51ad7884c5eb69c11d260a31da7a745e6b78e2 \
  -e DRONE_COMMIT_BRANCH=master \
  -e DRONE_BUILD_NUMBER=43 \
  -e DRONE_BUILD_STATUS=success \
  -w /drone/src \
  -v $(pwd):/drone/src \
  kitmi/tencent-cos-plugin
```
