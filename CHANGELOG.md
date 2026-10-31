# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.4.0] - 2026-10-31

### Added
- Arch Linux package (`markmirror-<version>-1-x86_64.pkg.tar.zst`) attached to every release; pacman installs the GTK/WebKitGTK dependencies.
- Automatic updates for MarkMirror installed from the Arch Linux package: the app downloads the new package, verifies it and installs it with pacman after asking for the password, then restarts.

## [2.3.1] - 2026-10-23

### Changed
- The "Commit e push" toolbar button is now always visible: it is enabled only when the open file is inside a Git repository, and disabled (with an explanatory tooltip) otherwise.

## [2.3.0] - 2026-10-22

### Added
- File → "Commit e push..." (Ctrl+Shift+G) and a toolbar button, shown when the open file is inside a Git repository: commits only the open file with a message of your choice, then pushes it to the remote.
- Unsaved changes are saved before committing; if the file has nothing to commit, a push-only option is offered.
- Branches without an upstream are pushed to the repository's remote and set to track it.

## [2.2.0] - 2026-10-21

### Changed
- The window title now shows the full path of the open file (e.g. `C:\Users\…\notes\README.md - MarkMirror`), so files with the same name in different folders can be told apart.

## [2.1.0] - 2026-10-19

### Added
- Help → "Verifica aggiornamenti" checks GitHub for a newer version, shows the release notes and installs it with one click.
- Automatic update check at startup, which can be turned off in Options → "Controlla aggiornamenti all'avvio".
- Downloaded updates are verified (size and SHA-256) before installing; unsaved changes are offered for saving first.
- Linux packages: .deb (Debian/Ubuntu) and .rpm (Fedora), in addition to the .tar.gz archive; MarkMirror appears in software centers.
- Linux: in-app updates through the administrator password prompt, and `sudo markmirror-update.sh` from the terminal.

### Changed
- New Windows installer: language choice (Italian/English), optional desktop and Start menu shortcuts, per-user installation without administrator rights; it offers to remove the old 2.0.0 installation.
- Release files now always include the version in their name (e.g. `MarkMirror-2.1.0-amd64-portable.exe`).
