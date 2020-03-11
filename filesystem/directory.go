package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// Directory represents a directory in the filesystem where content files are
// stored
type Directory struct {
	location string
}

// NewDirectory initializes a new Directory struct.
func NewDirectory(location string) *Directory {
	return &Directory{
		location: location,
	}
}

// getPath builds the path to a conversation content file in the directory.
func (d *Directory) getPath(conversationID int64) string {
	return path.Join(d.location, fmt.Sprintf("%d.html", conversationID))
}

// Create creates a content file for the given conversation ID.
func (d *Directory) Create(conversationID int64) error {
	filePath := d.getPath(conversationID)
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

// ReadFile reads the content file for the given conversation ID. If the file
// does not exist, it will return an error.
func (d *Directory) ReadFile(conversationID int64) ([]byte, error) {
	filePath := d.getPath(conversationID)
	if _, err := os.Stat(filePath); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(filePath)
	return data, err
}

// WriteFile overwrites the content file for the given conversation ID with the
// the given bytes.
func (d *Directory) WriteFile(conversationID int64, b []byte) error {
	filePath := d.getPath(conversationID)
	if _, err := os.Stat(filePath); err != nil {
		return err
	}

	err := ioutil.WriteFile(filePath, b, 0)
	return err
}

// Remove deletes the content file for the given conversation ID.
func (d *Directory) Remove(conversationID int64) error {
	filePath := d.getPath(conversationID)
	err := os.Remove(filePath)
	return err
}
