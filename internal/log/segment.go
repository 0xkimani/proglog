package log

// Segement keeps pointers to store and index fields for our segement
// to call its store and index files. next and base offsets use to know
// what offset to append new records under and to calculate the relative offsets
// for index entries. config used to compare the store file and index sizes to
// the configured limits, letting us know when the segment is max.
type segment struct {
	store                  *store
	index                  *index
	baseOffset, nextOffset uint64
	config                 Config
}

//
