package storage

import (
	"errors"
	"fmt"
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

func LoadEpisodesFromDisk(location string) ([]naming.ParseResult, error) {
	if location == "" {
		return nil, errors.New("Empty location given.")
	}
	mediaFiles, err := MediaFilesInDir(location)
	res := make([]naming.ParseResult, len(mediaFiles))

	if _, ok := err.(*os.PathError); ok {
		fmt.Printf("No path %s\n", location)
	}

	if err != nil {
		return res, err
	}

	np := naming.NewNameParser(location)
	for i, f := range mediaFiles {
		res[i] = np.Parse(f)
	}

	return res, nil
}
