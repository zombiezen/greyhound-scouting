package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

// Imagestore holds images for teams.
type Imagestore interface {
	HasTeamImage(num int) bool
	TeamImageURL(num int) (*url.URL, error)
	OpenTeamImage(num int) (io.ReadCloser, error)
}

const imageNameFormat = "%d.jpg"

type directoryImagestore struct {
	RootDir string
	RootURL *url.URL
}

func (store directoryImagestore) baseName(num int) string {
	return fmt.Sprintf(imageNameFormat, num)
}

func (store directoryImagestore) filePath(num int) string {
	return filepath.Join(store.RootDir, store.baseName(num))
}

func (store directoryImagestore) HasTeamImage(num int) bool {
	st, err := os.Stat(store.filePath(num))
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeType == 0
}

func (store directoryImagestore) TeamImageURL(num int) (*url.URL, error) {
	if !store.HasTeamImage(num) {
		return nil, StoreNotFound
	}
	return store.RootURL.Parse(store.baseName(num))
}

func (store directoryImagestore) OpenTeamImage(num int) (io.ReadCloser, error) {
	f, err := os.Open(store.filePath(num))
	if os.IsNotExist(err) {
		return nil, StoreNotFound
	}
	return f, err
}

// ReadTeamImage opens a team image and decodes it.
func ReadTeamImage(store Imagestore, num int) (image.Image, error) {
	f, err := store.OpenTeamImage(num)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}
