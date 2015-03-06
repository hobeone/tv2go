package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/hobeone/tv2go/quality"
	"github.com/hobeone/tv2go/types"
)

//Config is the base struct for Tv2Go configuration information.
type Config struct {
	DB            dbConfig
	WebServer     webConfig
	Storage       storageConfig
	MediaDefaults mediaDefaults
	Providers     []providerConfig
}

type providerConfig struct {
	Name string
	API  string
}

type webConfig struct {
	ListenAddress string // eg localhost:7000 or 0.0.0.0:8000
	EnableAPI     bool
}

type dbConfig struct {
	Path          string
	Verbose       bool   // turn on verbose db logging
	UpdateDb      bool   // if we should update db items
	WatchInterval int64  // how often to check for shows to update
	Type          string // file or memory (for testing)
}

type mediaDefaults struct {
	ShowQuality   quality.Quality
	EpisodeStatus types.EpisodeStatus
}

type storageConfig struct {
	Directories      []string
	NZBBlackhole     string
	TorrentBlackhole string
}

// NewConfig returns a Config struct with reasonable defaults set.
func NewConfig() *Config {
	return &Config{
		WebServer: webConfig{
			ListenAddress: "localhost:7000",
			EnableAPI:     false,
		},
		DB: dbConfig{
			Verbose:       true,
			UpdateDb:      true,
			WatchInterval: 60,
			Type:          "file",
		},
		Storage: storageConfig{
			Directories: []string{
				replaceTildeInPath("~/tv2go/storage"),
			},
			NZBBlackhole:     replaceTildeInPath("~/tv2go/nzb_blackhole"),
			TorrentBlackhole: replaceTildeInPath("~/tv2go/torrent_blackhole"),
		},
		MediaDefaults: mediaDefaults{
			ShowQuality:   quality.HDTV,
			EpisodeStatus: types.SKIPPED,
		},
	}
}

//NewTestConfig returns a Config instance suitable for use in testing.
func NewTestConfig() *Config {
	c := NewConfig()
	c.DB.Type = "memory"
	c.DB.Verbose = false
	return c
}

func replaceTildeInPath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	return strings.Replace(path, "~", dir, 1)
}

// ReadConfig Decodes and evaluates a json config file, watching for include cycles.
func (c *Config) ReadConfig(configPath string) error {
	absConfigPath, err := filepath.Abs(replaceTildeInPath(configPath))
	if err != nil {
		return fmt.Errorf("failed to expand absolute path for %s", configPath)
	}

	var f *os.File
	if f, err = os.Open(absConfigPath); err != nil {
		return fmt.Errorf("failed to open config: %v", err)
	}
	defer f.Close()

	filecont, err := ioutil.ReadAll(f)

	if err = json.Unmarshal(filecont, c); err != nil {
		extra := ""
		if serr, ok := err.(*json.SyntaxError); ok {
			if _, serr := f.Seek(0, os.SEEK_SET); serr != nil {
				fmt.Printf("seek error: %v\n", serr)
			}
			line, col, highlight := highlightBytePosition(f, serr.Offset)
			extra = fmt.Sprintf(":\nError at line %d, column %d (file offset %d):\n%s",
				line, col, serr.Offset, highlight)
		}
		return fmt.Errorf("error parsing JSON object in config file %s%s\n%v",
			f.Name(), extra, err)
	}

	c.DB.Path = replaceTildeInPath(c.DB.Path)
	for i, p := range c.Storage.Directories {
		c.Storage.Directories[i] = replaceTildeInPath(p)
	}
	c.Storage.NZBBlackhole = replaceTildeInPath(c.Storage.NZBBlackhole)

	return nil
}

// HighlightBytePosition takes a reader and the location in bytes of a parse
// error (for instance, from json.SyntaxError.Offset) and returns the line, column,
// and pretty-printed context around the error with an arrow indicating the exact
// position of the syntax error.
//
// Lifted from camlistore
func highlightBytePosition(f io.Reader, pos int64) (line, col int, highlight string) {
	line = 1
	br := bufio.NewReader(f)
	lastLine := ""
	thisLine := new(bytes.Buffer)
	for n := int64(0); n < pos; n++ {
		b, err := br.ReadByte()
		if err != nil {
			break
		}
		if b == '\n' {
			lastLine = thisLine.String()
			thisLine.Reset()
			line++
			col = 1
		} else {
			col++
			thisLine.WriteByte(b)
		}
	}
	if line > 1 {
		highlight += fmt.Sprintf("%5d: %s\n", line-1, lastLine)
	}
	highlight += fmt.Sprintf("%5d: %s\n", line, thisLine.String())
	highlight += fmt.Sprintf("%s^\n", strings.Repeat(" ", col+5))
	return
}
