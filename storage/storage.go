package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/naming"
)

// Broker is the interface between tv2go and the file system
type Broker struct {
	RootDirs []string
}

// NewBroker returns a pointer to a new Broker instance.
//
// Directories are turned into absolute paths if they are not already.
func NewBroker(dirs ...string) (*Broker, error) {
	glog.Infof("Creating new Broker with dirs: %v", dirs)
	b := &Broker{
		RootDirs: make([]string, len(dirs)),
	}
	for i := range dirs {
		abspath, err := filepath.Abs(dirs[i])
		if err != nil {
			return b, fmt.Errorf("Couldn't make path '%s' an absolute path: %s", dirs[i], err.Error())
		}
		b.RootDirs[i] = abspath
	}
	return b, nil
}

func dirInDirs(dir string, dirs []string) bool {
	for _, d := range dirs {
		if strings.HasPrefix(filepath.Clean(dir), filepath.Clean(d)) {
			return true
		}
	}
	return false
}

//CreateDir creates the given directory if it is under one of the Broker's RootDirs.
func (b *Broker) CreateDir(showdir string) (string, error) {
	showdir = filepath.Clean(showdir)
	if !filepath.IsAbs(showdir) {
		return "", fmt.Errorf("Non absolute directory given: %s", showdir)
	}
	if !dirInDirs(showdir, b.RootDirs) {
		return "", fmt.Errorf("Requested dir '%s' is not under any known root directory: %v", showdir, b.RootDirs)
	}

	//TODO: mask should be a broker attr
	err := os.MkdirAll(showdir, 0755)
	if err != nil {
		glog.Exitf("Error creating directory %s: %s", showdir, err.Error())
		return "", err
	}
	glog.Infof("Created directory %s", showdir)
	return showdir, nil
}

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

// LoadEpisodesFromDisk scans a directory for media files (as identified by naming.IsMediaFile)
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
