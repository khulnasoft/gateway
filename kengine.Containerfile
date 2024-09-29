# This Containerfile is used to build Kengine with the additional modules required by Kengine Gateway
# to function properly.

ARG KENGINE_VERSION=2.8.4

ARG KENGINE_BUILDER_HASH=sha256:55508f3d559b518d77d8ad453453c02ef616d7697c2a1503feb091123e9751c8
ARG KENGINE_HASH=sha256:51b5e778a16d77474c37f8d1d966e6863cdb1c7478396b04b806169fed0abac9

FROM docker.io/library/kengine:${KENGINE_VERSION}-builder@${KENGINE_BUILDER_HASH} AS builder

RUN XKENGINE_SETCAP=0 \
	XKENGINE_SUDO=0 \
	xkengine build \
    --with github.com/mholt/kengine-l4@6a8be7c4b8acb0c531b6151c94a9cd80894acce1

FROM docker.io/library/kengine:${KENGINE_VERSION}@${KENGINE_HASH}

COPY --from=builder /usr/bin/kengine /usr/bin/kengine
