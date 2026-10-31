#!/bin/sh
# Installs a MarkMirror release for Linux.
#
#   sudo markmirror-update.sh                  download the latest GitHub release (.tar.gz),
#                                              verify it with SHA256SUMS.txt and install it
#   sudo markmirror-update.sh FILE SHA256      install a release file already downloaded
#                                              (.deb, .rpm, .pkg.tar.zst or .tar.gz) after checking its SHA-256
#   sudo markmirror-update.sh FILE             same, without checksum (prints a warning)
#
# MarkMirror runs it through pkexec with the file it has downloaded and its
# SHA-256. The file is copied to a folder only root can access and checked
# again there, so it cannot be swapped between the check and the install.
set -e

REPO=alexis-82/markmirror

if [ "$(id -u)" -ne 0 ]; then
    echo "Run as root: sudo $0 $*" >&2
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

# mktemp -d creates a folder owned by root with mode 0700.
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT INT TERM

download() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL --proto '=https' --tlsv1.2 -o "$2" "$1"
    elif command -v wget >/dev/null 2>&1; then
        wget -q --https-only -O "$2" "$1"
    else
        echo "curl or wget is required." >&2
        exit 1
    fi
}

if [ -n "$1" ]; then
    if [ ! -f "$1" ]; then
        echo "File not found: $1" >&2
        exit 1
    fi
    NAME=$(basename "$1")
    FILE="$WORK/$NAME"
    cp "$1" "$FILE"
    if [ -n "$2" ]; then
        EXPECTED=$(printf '%s' "$2" | tr 'A-F' 'a-f')
        ACTUAL=$(sha256sum "$FILE" | cut -d ' ' -f 1)
        if [ "$ACTUAL" != "$EXPECTED" ]; then
            echo "SHA-256 mismatch for $NAME: nothing installed" >&2
            exit 1
        fi
    else
        echo "Warning: checksum of $NAME not verified" >&2
    fi
else
    download "https://api.github.com/repos/$REPO/releases/latest" "$WORK/latest.json"
    TAG=$(grep -o '"tag_name": *"[^"]*"' "$WORK/latest.json" | head -n 1 | sed 's/.*"\([^"]*\)"$/\1/')
    if ! printf '%s' "$TAG" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
        echo "Unexpected latest release tag: '$TAG'" >&2
        exit 1
    fi
    NAME="MarkMirror-${TAG#v}-linux-$ARCH.tar.gz"
    BASE="https://github.com/$REPO/releases/download/$TAG"
    FILE="$WORK/$NAME"
    echo "Downloading $BASE/$NAME"
    download "$BASE/$NAME" "$FILE"
    download "$BASE/SHA256SUMS.txt" "$WORK/SHA256SUMS.txt"
    EXPECTED=$(awk -v n="$NAME" '$2 == n { print $1 }' "$WORK/SHA256SUMS.txt")
    ACTUAL=$(sha256sum "$FILE" | cut -d ' ' -f 1)
    if [ -z "$EXPECTED" ] || [ "$ACTUAL" != "$EXPECTED" ]; then
        echo "SHA-256 of $NAME does not match SHA256SUMS.txt: nothing installed" >&2
        exit 1
    fi
fi

case "$NAME" in
    *.deb)
        # apt also installs any new dependency of the package.
        DEBIAN_FRONTEND=noninteractive apt-get install -y "$FILE"
        ;;
    *.rpm)
        if ! command -v dnf >/dev/null 2>&1; then
            echo "dnf not found: install $NAME with your package manager" >&2
            exit 1
        fi
        dnf install -y "$FILE"
        ;;
    *.pkg.tar.zst)
        if ! command -v pacman >/dev/null 2>&1; then
            echo "pacman not found: install $NAME with your package manager" >&2
            exit 1
        fi
        # pacman -U also installs any new dependency from the repositories.
        pacman -U --noconfirm "$FILE"
        ;;
    *.tar.gz)
        mkdir "$WORK/pkg"
        tar -xzf "$FILE" -C "$WORK/pkg" --no-same-owner
        if [ ! -f "$WORK/pkg/build/bin/MarkMirror" ] || [ ! -f "$WORK/pkg/build/linux/install.sh" ]; then
            echo "Invalid package: build/bin/MarkMirror or build/linux/install.sh missing" >&2
            exit 1
        fi
        (cd "$WORK/pkg" && sh build/linux/install.sh)
        ;;
    *)
        echo "Unsupported file: $NAME" >&2
        exit 1
        ;;
esac
