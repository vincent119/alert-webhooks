# Docker Deployment Guide

This guide explains how to run a fixed Alert Webhooks release from the public
Docker Hub repository. It covers a single-host Docker deployment, health
verification, routine operations, rollback, and the public image contract.

For Kubernetes deployments, use the maintained
[Helm guide](../../kubernetes/helm/alert-webhooks-chart/README.en.md) or
[Kustomize guide](../../kubernetes/kustomize/README.en.md).

## Prerequisites

- Docker Engine with access to Docker Hub.
- A local Alert Webhooks configuration file readable by UID `1500`.
- Port `9999` available on the host.
- A repository checkout when using the bundled
  `examples/config.expamle.yaml` example.

Do not store real tokens or passwords in the repository, image, shell history,
or CI logs. See the [configuration guide](./config_guide.md) for supported
settings and environment-variable overrides.

## Public image contract

- Repository: `docker.io/vincent119/alert-webhooks`.
- Current release tag: `v2.0.8`.
- Platforms: `linux/amd64` and `linux/arm64`.
- Runtime user: non-root UID `1500`.
- Health endpoint: `/healthy` on port `9999`.
- Version tags are immutable; Docker `latest` is not published.
- Docker Hub is the public distribution source. The private ECR production
  workflow remains unchanged.

The `v2.0.8` manifest digest is:

```text
sha256:2729da3594959d0d3c753e5d5c58f084000673e92f08859e70fafb1eb5da3ae1
```

Release evidence is available in the
[v2.0.8 GitHub Release](https://github.com/vincent119/alert-webhooks/releases/tag/v2.0.8).

## Start the container

Run these commands from the repository root when using the bundled example
configuration:

```bash
image="docker.io/vincent119/alert-webhooks:v2.0.8"
config_file="$(pwd)/examples/config.expamle.yaml"

docker pull "${image}"
docker run --detach \
  --name alert-webhooks \
  --publish 127.0.0.1:9999:9999 \
  --volume "${config_file}:/app/configs/config.production.yaml:ro" \
  --env APP_ENV=production \
  "${image}"
```

For a configuration outside the repository, replace `config_file` with its
absolute path. Review and secure the configuration before enabling notification
integrations.

## Verify the deployment

Check the public health endpoint:

```bash
curl --fail http://127.0.0.1:9999/healthy
```

A successful response has HTTP status `200`, `status` set to `ok`, and the
version loaded from the configuration file.

Confirm that the process runs as the expected non-root user:

```bash
docker exec alert-webhooks id -u
```

Expected output:

```text
1500
```

Inspect the Docker health status and application logs:

```bash
docker inspect --format '{{.State.Health.Status}}' alert-webhooks
docker logs --follow alert-webhooks
```

Stop the procedure and inspect logs if the health endpoint does not return
success or the UID is not `1500`.

## Routine operations

### View logs

```bash
docker logs --since 2h alert-webhooks
```

### Stop and start

```bash
docker stop alert-webhooks
docker start alert-webhooks
```

### Replace the container with another approved version

1. Record the current image tag and configuration path.
2. Pull the approved immutable version tag.
3. Stop and remove the existing container.
4. Run the new version with the same explicit settings.
5. Repeat the health and UID verification.

Removing the container does not remove the mounted configuration file. Confirm
that required logs are stored outside the container before replacement.

## Rollback

Because release tags are immutable, rollback uses the previously approved tag:

1. Stop and remove the unsuccessful container.
2. Run the previous image tag with the same configuration.
3. Verify `/healthy`, UID `1500`, and application logs.
4. Keep the unsuccessful version available for investigation; do not overwrite
   or reuse its tag.

## Troubleshooting

### Container does not start

```bash
docker logs alert-webhooks
docker inspect alert-webhooks
```

Confirm that the mounted file exists and is readable by UID `1500`:

```bash
ls -l "${config_file}"
```

### Health check fails

```bash
docker exec alert-webhooks \
  curl --fail http://127.0.0.1:9999/healthy
```

Check that the service is listening on port `9999` and that the mounted file is
available as `/app/configs/config.production.yaml`.

### Host port is already in use

Choose another loopback port while keeping the container port unchanged:

```bash
--publish 127.0.0.1:19999:9999
```

The corresponding host health URL becomes
`http://127.0.0.1:19999/healthy`.

## Local development image

Local builds should use a clearly local tag and must not be confused with the
published release:

```bash
docker build --tag alert-webhooks:dev .
```

The public release workflow is defined in
[publish-container-image.yml](../../.github/workflows/publish-container-image.yml).
Only a valid GitHub `v<major>.<minor>.<patch>` tag or an authorized manual
dispatch can publish an image. Ordinary branch pushes, pull requests, and
merges do not publish to Docker Hub.

## Related documentation

- [Configuration guide](./config_guide.md)
- [Kubernetes environment variables](./kubernetes-env-vars.md)
- [Helm installation](../../kubernetes/helm/alert-webhooks-chart/README.en.md)
- [Kustomize installation](../../kubernetes/kustomize/README.en.md)
