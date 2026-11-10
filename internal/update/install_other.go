//go:build !windows && !linux

package update

import (
	"errors"
	goruntime "runtime"
)

func Install(path, sum string) error {
	return errors.New("installazione automatica non disponibile su questo sistema: scarica la nuova versione da GitHub")
}

func PlatformAsset(assets []Asset, version string) (Asset, error) {
	return PickAsset(assets, version, goruntime.GOOS, goruntime.GOARCH)
}
