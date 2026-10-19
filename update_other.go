//go:build !windows && !linux

package main

import (
	"errors"
	goruntime "runtime"

	"markmirror-go/internal/update"
)

func installUpdate(path, sum string) error {
	return errors.New("installazione automatica non disponibile su questo sistema: scarica la nuova versione da GitHub")
}

func platformAsset(assets []update.Asset, version string) (update.Asset, error) {
	return update.PickAsset(assets, version, goruntime.GOOS, goruntime.GOARCH)
}
