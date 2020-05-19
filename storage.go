package rubik

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Storage is the Container Access of your storage/ folder
var Storage = StorageContainer{
	path: filepath.Join(".", "storage"),
}

// GetStorageContainers returns the names of containers present in your
// storage/ folder. You can access them by calling `Storage.Access`
// API and use Get or Put to work with your files.
func GetStorageContainers() []string {
	containers := []string{}
	base := filepath.Join(".", "storage")
	files, err := ioutil.ReadDir(base)
	if err != nil {
		return containers
	}

	for _, f := range files {
		if f.IsDir() {
			containers = append(containers, f.Name())
		}
	}

	return containers
}

// StorageContainer is abstracted struct to access your
// storage files. It can access of remove a whole container.
//
// Container corresponds to a single directory in your storage
// folder and will have access to files only inside this
// container/directory
type StorageContainer struct {
	path string
}

// FileStore lets you perform CRUD on files of StorageContainer
// FileStore returns the name of container you are accessing
// by Name field
type FileStore struct {
	Name     string
	fullPath string
}

// Access a FileStore from your StorageContainer.
// It can be viewed as accessing a specific folder inside
// your storage/ folder and performing operations inside
// of that folder
func (s StorageContainer) Access(name string) (FileStore, error) {
	storeFolder := filepath.Join(s.path, name)
	if f, _ := os.Stat(storeFolder); f == nil {
		return FileStore{}, errors.New("No such container.")
	}
	return FileStore{Name: name, fullPath: storeFolder}, nil
}

// Remove a FileStore from your StorageContainer.
// Removing a FileStore will remove all the files inside
// the FileStore
func (s StorageContainer) Remove(name string) error {
	storeFolder := filepath.Join(s.path, name)
	// remove the directory named 'name'
	if f, _ := os.Stat(storeFolder); f != nil {
		return os.RemoveAll(storeFolder)
	}
	return nil
}

// Get a file from this FileStore, returs byte slice
func (fs FileStore) Get(file string) []byte {
	outFile := filepath.Join(fs.fullPath, file)
	b, err := ioutil.ReadFile(outFile)
	if err != nil {
		return nil
	}
	return b
}

// Put a file inside this FileStore given the content
// as parameter
func (fs FileStore) Put(file string, content []byte) error {
	inFile := filepath.Join(fs.fullPath, file)
	if f, _ := os.Stat(inFile); f != nil {
		os.Remove(inFile)
	}
	return ioutil.WriteFile(inFile, content, 0755)
}

// Delete a file from the FileStore, returns error
func (fs FileStore) Delete(file string) error {
	delFile := filepath.Join(fs.fullPath, file)
	return os.Remove(delFile)
}

// Has checks if the file by given name is present
// inside this FileStore
func (fs FileStore) Has(file string) bool {
	if f, _ := os.Stat(filepath.Join(fs.fullPath, file)); f != nil {
		return true
	}
	return false
}
