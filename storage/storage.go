package storage

import (
	"os"
	"path/filepath"

	"github.com/hobeone/tv2go/naming"
)

// MediaFilesInDir returns a list of all media files in the given directory. An
// error is returned if there was problem walking the directory.
func MediaFilesInDir(directory string) ([]string, error) {
	var mediaFiles []string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if naming.IsMediaFile(path) {
			mediaFiles = append(mediaFiles, path)
		}
		return nil
	}
	err := filepath.Walk(directory, walkFn)
	if err != nil {
		return nil, err
	}
	return mediaFiles, nil
}

/*
func LoadEpisodesFromDisk() error {

	mediaFiles, err := MediaFilesInDir(dbshow.Location)
	if err != nil {
		return err
	}

	for _, f := range mediaFiles {
		np := name_parser.New(f)
	}

	return nil
}
*/
