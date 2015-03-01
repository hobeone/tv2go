package storage

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
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

func TestSaveToFile(t *testing.T) {
	RegisterTestingT(t)
	//flag.Set("logtostderr", "true")
	testdir, err := ioutil.TempDir("testdata", "testing")
	if err != nil {
		t.Fatalf("Couldn't create a tempdir for testing: %s", err)
	}
	defer os.RemoveAll(testdir)

	b, err := NewBroker(testdir)
	Expect(err).ToNot(HaveOccurred())

	savedname, err := b.SaveToFile(testdir, "testFiLeName", []byte("teststring"))
	Expect(err).ToNot(HaveOccurred())
	Expect(savedname).To(ContainSubstring("testFiLeName"))

	// Empty filename
	savedname, err = b.SaveToFile(testdir, "", []byte("teststring"))
	Expect(err).ToNot(HaveOccurred())
	Expect(savedname).To(ContainSubstring("unknown"))
}

func TestFileReadable(t *testing.T) {
	RegisterTestingT(t)
	//flag.Set("logtostderr", "true")

	testdir, err := ioutil.TempDir("testdata", "testing")
	if err != nil {
		t.Fatalf("Couldn't create a tempdir for testing: %s", err)
	}
	defer os.RemoveAll(testdir)

	testfilepath := testdir + "testReadableFile"

	err = ioutil.WriteFile(testfilepath, []byte("testfile\n"), 0644)
	Expect(err).ToNot(HaveOccurred())

	b, err := NewBroker(testdir)
	err = b.FileReadable(testfilepath)
	Expect(err).ToNot(HaveOccurred())
}

func TestMoveFile(t *testing.T) {
	RegisterTestingT(t)

	testdir, err := ioutil.TempDir("testdata", "testing")
	if err != nil {
		t.Fatalf("Couldn't create a tempdir for testing: %s", err)
	}
	defer os.RemoveAll(testdir)

	b, err := NewBroker(testdir)
	Expect(err).ToNot(HaveOccurred())

	testfilepath, err := b.SaveToFile(testdir, "testfileorig", []byte("test\n"))
	Expect(err).ToNot(HaveOccurred())
	testfilepathdest := filepath.Join(testdir, "testFileDest")

	err = b.MoveFile(testfilepath, testfilepathdest)
	Expect(err).To(MatchError("Both source and destination must be absolute paths"))

	testfilepath, err = filepath.Abs(testfilepath)
	Expect(err).ToNot(HaveOccurred())
	testfilepathdest, err = filepath.Abs(testfilepathdest)
	Expect(err).ToNot(HaveOccurred())
	err = b.MoveFile(testfilepath, testfilepathdest)
	Expect(err).ToNot(HaveOccurred())

	err = b.FileReadable(testfilepathdest)
	Expect(err).ToNot(HaveOccurred())
}

// Bleh, not sure how to test this reliably
func TestMoveAcrossFilesystem(t *testing.T) {
	RegisterTestingT(t)
	flag.Set("logtostderr", "true")

	testdir, err := ioutil.TempDir("testdata", "testing")
	if err != nil {
		t.Fatalf("Couldn't create a tempdir for testing: %s", err)
	}
	defer os.RemoveAll(testdir)

	testdir2, err := ioutil.TempDir("/tmp", "testing")
	if err != nil {
		t.Fatalf("Couldn't create a tempdir for testing: %s", err)
	}
	defer os.RemoveAll(testdir2)
	b, err := NewBroker(testdir)

	testfilepath, err := b.SaveToFile(testdir, "testfileorig", []byte("test\n"))
	if err != nil {
		t.Fatalf("Couldn't write file: %s", err)
	}
	testfilepath, err = filepath.Abs(testfilepath)
	testfilepathdest := filepath.Join(testdir2, "testFileDest")

	spew.Dump(testfilepath)
	spew.Dump(testfilepathdest)
	err = b.MoveFile(testfilepath, testfilepathdest)
	Expect(err).ToNot(HaveOccurred())

	err = b.FileReadable(testfilepathdest)
	Expect(err).ToNot(HaveOccurred())
}
