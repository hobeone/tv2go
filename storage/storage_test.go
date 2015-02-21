package storage

import (
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
