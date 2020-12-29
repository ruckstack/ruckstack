# Release Process

Most people don't care how the sausage is made. But if you are interested, this is what we do:

1. Write release post in https://github.com/ruckstack/ruckstack.github.io
1. Ensure any needed docs are created/updated in ruckstack.github.io   
1. Download built artifacts from main branch https://github.com/ruckstack/ruckstack/actions?query=branch%3Amain
1. Publish new release at https://github.com/ruckstack/ruckstack/releases with tag version vX.Y.Z
1. Update download page in ruckstack.github.io
1. Commit and push ruckstack.github.io
1. Make sure site built and download links work
1. Push docker images (below)
1. Update version in constants.go and BUILD.sh and commit
1. Tell the world
   
## Push docker images

```
export VERSION="vX.Y.Z"
docker pull ghcr.io/ruckstack/ruckstack:snapshot-main

docker tag ghcr.io/ruckstack/ruckstack:snapshot-main ghcr.io/ruckstack/ruckstack:${VERSION}
docker tag ghcr.io/ruckstack/ruckstack:snapshot-main ruckstack/ruckstack:${VERSION}
docker tag ghcr.io/ruckstack/ruckstack:snapshot-main ghcr.io/ruckstack/ruckstack:latest
docker tag ghcr.io/ruckstack/ruckstack:snapshot-main ruckstack/ruckstack:latest

docker push ghcr.io/ruckstack/ruckstack:${VERSION}
docker push ghcr.io/ruckstack/ruckstack:latest
docker push ruckstack/ruckstack:${VERSION}
docker push ruckstack/ruckstack:latest
```
