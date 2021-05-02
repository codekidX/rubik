package rubik

import (
	"os"
	"path/filepath"
	"testing"
)

// Requirements: (scaffold script)
// 1. test/ folder inside storage folder
// 2. testFile inside test/ folder

func TestGetStorageContainers(t *testing.T) {
	names := GetStorageContainers()
	if len(names) != 1 {
		t.Errorf("There should be only test folder inside container")
	}

	if names[0] != "test" {
		t.Errorf("Test folder name is not in the list: %v", names)
	}
}

func TestAccess(t *testing.T) {
	_, err := Storage.Access("test")
	if err != nil {
		t.Error(err)
	}

	// if this dir is present remove it
	testbP := filepath.Join(".", "storage", "testb")
	os.RemoveAll(testbP)
	// test case when folder not present -- test if Access creates it
	_, err = Storage.Access("testb")
	if err != nil {
		t.Error(err)
	}

	if f, _ := os.Stat(testbP); f == nil {
		t.Error("Access() did not create a folder")
	}
}

func TestRemove(t *testing.T) {
	err := Storage.Remove("testb")
	if err != nil {
		t.Error(err)
	}
}

func TestFsGet(t *testing.T) {
	fs, err := Storage.Access("test")
	if err != nil {
		t.Error(err)
	}

	fileb := fs.Get("testFile")
	if fileb == nil {
		t.Errorf("Did not read file inside the file store")
	} else if string(fileb) != "test" {
		t.Errorf("File content is read wrong: %s", string(fileb))
	}
}

func TestFsPut(t *testing.T) {
	fs, err := Storage.Access("test")
	if err != nil {
		t.Error(err)
	}

	testFilebP := filepath.Join(".", "storage", "test", "testFileb")
	os.Remove(testFilebP)

	err = fs.Put("testFileb", []byte("testb"))
	if err != nil {
		t.Error(err)
	}

	if f, _ := os.Stat(testFilebP); f == nil {
		t.Error("FileStore.Put() did not write test/testFileb file")
	}
}

func TestFsHas(t *testing.T) {
	fs, err := Storage.Access("test")
	if err != nil {
		t.Error(err)
	}

	hasFile := fs.Has("testFileb")
	if !hasFile {
		t.Error("FileStore.Has() says testFileb not present; which was written in previous test")
	}
}

func TestFsDelete(t *testing.T) {
	fs, err := Storage.Access("test")
	if err != nil {
		t.Error(err)
	}

	err = fs.Delete("testFileb")
	if err != nil {
		t.Error(err)
	}
}
