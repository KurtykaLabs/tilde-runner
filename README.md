# Tilde Runner deployment

Public deployment configuration and measured-release workflows for Tilde Runner.
The Rust application source and image-build workflow remain in the private
`tildeapp/tilde` repository. The image is published privately to
`ghcr.io/tildeapp/tilde-runner`.

## Boot experiment

The initial `v0.1.0-m1.*` configuration runs the hosted service with disposable
installation credentials supplied through Tinfoil-managed secrets. It exposes
health, public identity, activation, and authenticated command endpoints.

This configuration uses a RAM-backed state directory. Stopping or replacing the
container loses its installation keys and unfinished work. Use only disposable
accounts and data. This configuration does not provide persistent enrollment,
attestation-gated private key custody, or the complete hosted chat migration.
Deployment is not approval to release household keys.

Create repository secrets `TILDE_RUNNER_CONFIG` and `TILDE_RUNNER_UNLOCK` through
Tinfoil's dashboard or CLI. The first contains the hosted runner configuration;
the second is a random 32-byte key encoded as 64 hexadecimal characters. Keep
both values out of this repository, GitHub Actions, and command-line arguments.
The private source repository defines the configuration schema.

Configure read-only GHCR credentials in Tinfoil before deployment. The minimum
measured allocation is two CPUs and 8192 MB RAM. The process runs as an
unprivileged user with a read-only root and a bounded writable `/state` tmpfs.

## Release procedure

1. In the private source repository, run **Publish Private Runner Image** against
   the reviewed source revision, or push a `runner-image-*` tag. The workflow runs
   the Docker build/tests and publishes a Linux amd64 image. Confirm package
   visibility is private and copy the immutable image reference from its summary.
2. Update and review `tinfoil-config.yml`. Keep credentials, private keys,
   account/family identifiers, and application source out of this repo.
   Secret names and image digests are public; secret values are not configuration.
3. Run **Tinfoil Release** on the reviewed commit with a new version. It creates a
   tag and dispatches **Tinfoil Release - Publish** on that tag. Wait for a
   successful GitHub release containing the deployment manifest and measurement.
   Creating a tag alone is insufficient.
4. Deploy that exact release using the Tinfoil dashboard or CLI. Leave debug and
   confidential-computing bypass disabled.
5. Verify the deployed endpoint against the expected repository and exact release
   using the Tinfoil SDK or CLI. An ordinary HTTPS health response is not evidence
   of attestation. Pin the approved measurement in client and keyserver policy
   before granting access to user keys or data.

Image publication and deployment release are separate, explicit operations. This
repository does not check out the private source or need a private-source token.
Public attestation pins an image; it does not make that image's source auditable.

## Upstream references

Workflows follow the [Tinfoil template](https://github.com/tinfoilsh/tinfoil-containers-template/tree/0eddc320b8f328d7a3c057152596934444ac2d75),
with commit-pinned actions and release preflight checks.

- [Configuration](https://docs.tinfoil.sh/containers/configuration)
- [Private images and registry credentials](https://docs.tinfoil.sh/containers/private-images)
- [Private secrets](https://docs.tinfoil.sh/containers/private-secrets)
