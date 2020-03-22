package filesystem

import (
	"log"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var dmp *diffmatchpatch.DiffMatchPatch = diffmatchpatch.New()

type File struct {
	content      string
	lastReadTime time.Time
}

type Update struct {
	ConversationID int64
	Patch          string
}

type CachedWriter struct {
	directory *Directory
	files     map[int64]File
	Write     chan *Update
}

func NewCachedWriter(directory *Directory) *CachedWriter {
	return &CachedWriter{
		directory: directory,
		files:     make(map[int64]File),
		Write:     make(chan *Update),
	}
}

func (cw *CachedWriter) Run() {
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
