The plugin upload files to your tencent cos bucket.

# Usage

The following settings changes this plugin's behavior.

* command: (optional) upload, download or delete, default as upload
* bucket: storage bucket name
* region: storage region
* key: storage key, should contain the trailing slash for folders
* localPath: local path

Below is an example `.drone.yml` that uses this plugin.

```yaml
kind: pipeline
name: default

steps:
- name: run kitmi/tencent-cos-plugin plugin
  image: kitmi/drone-plugin-tencent-cos
  pull: if-not-exists
  settings:
    command: download
    bucket: foo
    region: bar
    key: conf/sso
    localPath: app/app.conf
```

# Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t kitmi/drone-plugin-tencent-cos -f docker/Dockerfile .
```

# Testing

Execute the plugin from your current working directory:

```text
docker run --rm -e COS_SECRETID=<secretId> \
  -e COS_SECRETKEY=<secretKey> \
  -e PLUGIN_BUCKET=<bucket> \
  -e PLUGIN_REGION=<region> \
  -e PLUGIN_KEY=test \
  -e PLUGIN_LOCAL_PATH=release \
  -e DRONE_COMMIT_SHA=8f51ad7884c5eb69c11d260a31da7a745e6b78e2 \
  -e DRONE_COMMIT_BRANCH=master \
  -e DRONE_BUILD_NUMBER=43 \
  -e DRONE_BUILD_STATUS=success \
  -w /drone/src \
  -v $(pwd):/drone/src \
  kitmi/drone-plugin-tencent-cos
```
