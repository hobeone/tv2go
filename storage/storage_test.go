package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func TestMediaFilesFromDir(t *testing.T) {
	RegisterTestingT(t)

	files, err := MediaFilesInDir("testdata/testdir")
	Expect(err).ToNot(HaveOccurred())
	Expect(files).To(HaveLen(1))
	Expect(files).To(ContainElement("testdata/testdir/MediaFile.mkv"))
}

func TestNewBroker(t *testing.T) {
	RegisterTestingT(t)

	testdir, err := ioutil.TempDir("testdata", "testing")
	if err != nil {
		t.Fatalf("Couldn't create a tempdir for testing: %s", err)
	}
	abstestdir, err := filepath.Abs(testdir)
	if err != nil {
		t.Fatalf("Couldn't make an absoulte path of testdir: %s", err)
	}
	defer os.RemoveAll(abstestdir)

	b, err := NewBroker(testdir)
	if err != nil {
		t.Fatalf("Error creating broker: %s", err)
	}

	Expect(b.RootDirs[0]).To(Equal(abstestdir))

	b, err = NewBroker("")
	Expect(err).ToNot(HaveOccurred())
	absofempty, _ := filepath.Abs("")
	Expect(b.RootDirs[0]).To(Equal(absofempty))
}

func TestCreateShowDir(t *testing.T) {
	RegisterTestingT(t)

	testdir1, err := ioutil.TempDir("testdata", "testing")
	if err != nil {
		t.Fatalf("Couldn't create a tempdir for testing: %s", err)
	}
	testdir1, err = filepath.Abs(testdir1)
	if err != nil {
		t.Fatalf("Couldn't make an absoulte path of testdir: %s", err)
	}
	defer os.RemoveAll(testdir1)

	b, err := NewBroker(testdir1)
	if err != nil {
		t.Fatalf("Error creating broker: %s", err)
	}
	_, err = b.CreateDir("non-absolute-dir")
	Expect(err).To(HaveOccurred())

	_, err = b.CreateDir("/testdir3/showdir")
	Expect(err).To(HaveOccurred())

	fp := filepath.Join(testdir1, "showdir")
	newdir, err := b.CreateDir(fp)
	Expect(err).ToNot(HaveOccurred())
	Expect(newdir).To(Equal(fp))
}
