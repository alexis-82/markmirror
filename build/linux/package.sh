#!/bin/sh
# Creates build/bin/MarkMirror-<version>-linux-<arch>.tar.gz, the Linux file
# attached to a GitHub release. Run from the project root after the Linux
# build, e.g.
#   wails build -tags webkit2_41 -ldflags "-X main.version=v2.1.0"
#   VERSION=2.1.0 sh build/linux/package.sh
# The archive keeps the project layout expected by install.sh and
# markmirror-update.sh.
set -e

if [ -z "$VERSION" ]; then
    echo "VERSION not set (X.Y.Z, without the v of the tag)." >&2
    exit 1
fi
if [ ! -f build/bin/MarkMirror ]; then
    echo "build/bin/MarkMirror not found: run 'wails build -tags webkit2_41' from the project root first." >&2
    exit 1
fi

case "$(uname -m)" in
    x86_64) ARCH=amd64 ;;
    aarch64 | arm64) ARCH=arm64 ;;
    *)
        echo "Unsupported architecture: $(uname -m)" >&2
        exit 1
        ;;
esac

# Same name as update.AssetNames in internal/update/update.go.
OUT="build/bin/MarkMirror-$VERSION-linux-$ARCH.tar.gz"
tar -czf "$OUT" \
    build/bin/MarkMirror \
    build/linux/install.sh \
    build/linux/markmirror-launch.sh \
    build/linux/markmirror-update.sh \
    build/linux/markmirror.desktop \
    build/appicon.png
echo "Created $OUT"
