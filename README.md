# Tilde Runner deployment

Public deployment configuration and measured-release workflows for Tilde Runner.
The Rust application source and image-build workflow remain in the private
`KurtykaLabs/tilde` repository. The image is published privately to
`ghcr.io/kurtykalabs/tilde-runner`.

## Deployment status

No deployable release has been published. `tinfoil-config.yml.example` is an
incomplete template, deliberately not the root `tinfoil-config.yml` that Tinfoil
deploys. The release workflows refuse to proceed without that file.

The example pins a published private Linux amd64 prototype image. Its default
command verifies the inference provider and exits; it is not a hosted service.

Before the first release, finish the runner's hosted startup/pairing adapter and
private-secret delivery, publish its image, and commit a complete
`tinfoil-config.yml` with the actual immutable image digest, startup arguments,
network policy, and keyserver/volume configuration. A verification-only image
that exits is not a serving runner.

## Release procedure

1. In the private source repository, run **Publish Private Runner Image** against
   the reviewed source revision, or push a `runner-image-*` tag. The workflow runs
   the Docker build/tests and publishes a Linux amd64 image. Confirm package
   visibility is private and copy the immutable image reference from its summary.
2. Update and review this repository's complete `tinfoil-config.yml`. Keep credentials,
   private keys, account/family identifiers, and application source out of this repo.
   Secret names and image digests are public; secret values are not configuration.
3. Run **Tinfoil Release** on the reviewed commit with a new version such as
   `v0.1.0`. It creates a tag and dispatches **Tinfoil Release - Publish** on that
   tag. Wait for a successful GitHub release containing the deployment manifest
   and measurement. Creating a tag alone is insufficient.
4. Connect this repository to Tinfoil and select that measured release. Configure
   GHCR read-only registry credentials in Tinfoil separately. Private registry
   support must be enabled for the organization.
5. Approve and pin the release in the client and keyserver policy before using it.
   Deployment or measurement alone does not authorize access to user keys/data.

Image publication and deployment release are separate, explicit operations. This
repository does not check out the private source or need a private-source token.
Public attestation pins an image; it does not make that image's source auditable.

## Upstream references

Workflows follow the [Tinfoil template](https://github.com/tinfoilsh/tinfoil-containers-template/tree/0eddc320b8f328d7a3c057152596934444ac2d75),
with commit-pinned actions and release preflight checks.

- [Configuration](https://docs.tinfoil.sh/containers/configuration)
- [Private images and registry credentials](https://docs.tinfoil.sh/containers/private-images)
- [Private secrets](https://docs.tinfoil.sh/containers/private-secrets)
