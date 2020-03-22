package filesystem

import (
	"log"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var dmp *diffmatchpatch.DiffMatchPatch = diffmatchpatch.New()

// File represents a cached file.
type File struct {
	content      string
	lastReadTime time.Time
}

// Update represents a conversation content file update.
type Update struct {
	ConversationID int64
	Patch          string
}

// CachedWriter encapsulates the behaviour of updating files in the filesytem
// with caching capabilities to minimize I/O operations.
type CachedWriter struct {
	directory *Directory
	files     map[int64]File
	Write     chan *Update
}

// NewCachedWriter initializes a new CachedWriter.
func NewCachedWriter(directory *Directory) *CachedWriter {
	return &CachedWriter{
		directory: directory,
		files:     make(map[int64]File),
		Write:     make(chan *Update),
	}
}

// Run loops indefinitely and blocks on a channel where Update structs will come
// in and be processed one by one.
func (cw *CachedWriter) Run() {
	/* TODO: cache the file content so that we don't have to read and write the
	   file every time a new Update comes in. */

	for {
		select {
		case update := <-cw.Write:
			content, err := cw.directory.ReadFile(update.ConversationID)
			if err != nil {
				log.Printf("Failed to read content file: %v", err)
				continue
			}

			patches, err := dmp.PatchFromText(update.Patch)
			if err != nil {
				log.Printf("Could not process patch string: %s", update.Patch)
				continue
			}

			newContent, okList := dmp.PatchApply(patches, string(content))
			if !okList[0] {
				log.Printf("Could not apply patch: %s", update.Patch)
				continue
			}

			cw.directory.WriteFile(update.ConversationID, []byte(newContent))
			if err != nil {
				log.Printf("Failed to write content file: %v", err)
				continue
			}
		}
	}
}
