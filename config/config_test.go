package config

import (
	//	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestReadConfigFailsOnNonExistingPath(t *testing.T) {
	c := NewConfig()
	path := "/does/not/exist"
	err := c.ReadConfig(path)
	if err == nil {
		t.Errorf("Expected PathError on non existing path: %s", path)
	}
}

func TestReadConfigFailsOnBadFormat(t *testing.T) {
	c := NewConfig()
	path := "../testdata/configs/bad_config.json"
	err := c.ReadConfig(path)

	if err == nil {
		t.Error("Expected error on bad format config: ", path)
	}
}

func TestDefaultsGetOverridden(t *testing.T) {
	c := NewConfig()
	if c.Mail.UseSMTP {
		t.Fatal("Expected UseSMTP to be false")
	}
	path := "../testdata/configs/test_config.json"
	err := c.ReadConfig(path)
	if err != nil {
		t.Fatalf("Expected no errors when parsing: %s, got %s", path, err)
	}
	if !c.Mail.UseSMTP {
		t.Fatal("Expected c.Mail.UseSMTP to be true")
	}
}
