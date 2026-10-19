# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
