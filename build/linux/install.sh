#!/bin/sh
# Installs MarkMirror system-wide. Run from the project root as root after
# `wails build -tags webkit2_41`.
#   sudo sh build/linux/install.sh              install
#   sudo sh build/linux/install.sh --uninstall  remove
set -e

FILES="/usr/local/bin/markmirror
/usr/local/bin/markmirror-launch.sh
/usr/share/applications/markmirror.desktop
/usr/share/pixmaps/markmirror.png"

refresh_menus() {
    command -v update-desktop-database >/dev/null 2>&1 &&
        update-desktop-database -q /usr/share/applications || true
}

if [ "$1" = "--uninstall" ]; then
    echo "$FILES" | xargs rm -f
    refresh_menus
    echo "MarkMirror removed. Settings are kept in ~/.config/MarkMirror"
    exit 0
fi

if [ ! -f build/bin/MarkMirror ]; then
    echo "build/bin/MarkMirror not found: run 'wails build -tags webkit2_41' from the project root first." >&2
    exit 1
fi

install -m 0755 build/bin/MarkMirror /usr/local/bin/markmirror
install -m 0755 build/linux/markmirror-launch.sh /usr/local/bin/markmirror-launch.sh
install -m 0644 build/linux/markmirror.desktop /usr/share/applications/markmirror.desktop
install -m 0644 build/appicon.png /usr/share/pixmaps/markmirror.png
refresh_menus
echo "MarkMirror installed. Runtime dependency: libwebkit2gtk-4.1-0"
