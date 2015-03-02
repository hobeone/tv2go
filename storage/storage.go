package storage

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/naming"
	"github.com/termie/go-shutil"
)

// Broker is the interface between tv2go and the file system
type Broker struct {
	RootDirs []string
}

// NewBroker returns a pointer to a new Broker instance.
//
// Directories are turned into absolute paths if they are not already.
func NewBroker(dirs ...string) (*Broker, error) {
	if len(dirs) < 1 {
		return nil, fmt.Errorf("No directories given")
	}
	glog.Infof("Creating new Broker with dirs: %v", dirs)
	b := &Broker{
		RootDirs: make([]string, len(dirs)),
	}
	for i := range dirs {
		abspath, err := filepath.Abs(dirs[i])
		if err != nil {
			return b, fmt.Errorf("Couldn't make path '%s' an absolute path: %s", dirs[i], err.Error())
		}
		if abspath != dirs[i] {
			glog.Infof("Turned path '%s' into absolute path '%s'", dirs[i], abspath)
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

func (b *Broker) Readable(path string) bool {
	_, err := os.Stat(path)

	if err != nil {
		return false
	}
	return true
}

// FileReadable returns an error if path can't be stat'ed or is a directory
func (b *Broker) FileReadable(path string) error {
	fi, err := os.Stat(path)

	if err != nil {
		return err
	}

	if fi.IsDir() {
		return fmt.Errorf("%s is a directory not a file", path)
	}
	return nil
}

// MoveFile moves the src file to the dst path.  Both must be absolute paths.
// It will try to just rename the file from the src to dst name but if that
// fails it will copy the src to the dst and then remove the old.
func (b *Broker) MoveFile(src, dst string) error {
	if !filepath.IsAbs(src) || !filepath.IsAbs(dst) {
		return fmt.Errorf("Both source and destination must be absolute paths")
	}

	// Hella odd that Go doesn't have something like Python's shutil, oh well
	err := os.Rename(src, dst)

	if _, ok := err.(*os.LinkError); ok {
		glog.Infof("Rename failed: %s (are %s and %s on different filesystems?), falling back to copy and remove.", err, src, dst)
	} else {
		glog.Errorf("Rename returned an unknown error: %s", err)
		return err
	}

	_, err = shutil.Copy(src, dst, true)
	if err != nil {
		glog.Infof("Copy to '%s' failed: %s", err)
		return err
	}
	glog.Infof("Copy to '%s' successful.  Removing '%s'", dst, src)
	err = os.Remove(src)
	if err != nil {
		if os.IsNotExist(err) {
			glog.Infof("remove failed, source file was already removed: %s", err)
			return nil
		}
		return err
	}
	return nil
}

// SaveToFile saves the given content to the filename in the directory.  If
// fname is empty it will save to a temporary file.
func (b *Broker) SaveToFile(dirname, fname string, content []byte) (string, error) {
	joined := ""
	if fname == "" {
		fh, err := ioutil.TempFile(dirname, "unknown")
		if err != nil {
			return "", fmt.Errorf("Error creating file (and no filename was given): %s", err)
		}
		joined = fh.Name()
		fh.Close()
	} else {
		joined = filepath.Join(dirname, fname)
	}
	abspath, err := filepath.Abs(joined)
	if err != nil {
		return "", fmt.Errorf("Couldn't make an absolute path for %s", joined)
	}

	glog.Infof("Write content to: '%s'", abspath)
	err = ioutil.WriteFile(abspath, content, 0644)
	if err != nil {
		glog.Errorf("Error writing to file '%s': %s", abspath, err)
		return "", err
	}
	return abspath, nil
}
