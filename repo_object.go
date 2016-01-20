package git

import (
	"fmt"
	"io"
	"os"
)

// ObjectNotFound error returned when a repo query is performed for an ID that does not exist.
type ObjectNotFound ObjectID

func (id ObjectNotFound) Error() string {
	return fmt.Sprintf("object not found: %s", ObjectID(id))
}

type ObjectType int

const (
	ObjectCommit ObjectType = 0x10
	ObjectTree   ObjectType = 0x20
	ObjectBlob   ObjectType = 0x30
	ObjectTag    ObjectType = 0x40
)

type Object struct {
	Type ObjectType
	Size int64
	Data io.ReadCloser
}

func (t ObjectType) String() string {
	switch t {
	case ObjectCommit:
		return "commit"
	case ObjectTree:
		return "tree"
	case ObjectBlob:
		return "blob"
	default:
		return ""
	}
}

// Given a SHA1, find the pack it is in and the offset, or return nil if not
// found.
func (repo *Repository) findObjectPack(id ObjectID) (*idxFile, uint64) {
	for _, indexfile := range repo.indexfiles {
		if offset, ok := indexfile.offsetValues[id]; ok {
			return indexfile, offset
		}
	}
	return nil, 0
}

func (repo *Repository) getRawObject(id ObjectID, metaOnly bool) (*Object, error) {
	ot, length, dataRc, err := readObjectFile(filepathFromSHA1(repo.Path, id.String()), metaOnly)
	if err == nil {
		return &Object{ot, length, dataRc}, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	if pack, _ := repo.findObjectPack(id); pack == nil {
		return nil, ObjectNotFound(id)
	}
	pack, offset := repo.findObjectPack(id)
	ot, length, data, err := readObjectBytes(pack.packpath, &repo.indexfiles, offset, metaOnly)
	if err != nil {
		return nil, err
	}
	return &Object{ot, length, data}, nil
}
