package log

import (
	"io"
	"os"

	gommap "github.com/tysonmote/gommap"
)

// *Width constants define the number of bytes that make up
// each index entry.
//
// Index entries contain two fields: the record's offset and its position
// in the store file. Offsets and Positions are uint32 and uint64 so they take
// up to 4 and 8 bytes of space. We use the entWidth to jump straight to the position
// of an entry given its offset since the position in the file is offset * entWidth.
var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

// index defines our index file, which comprises a persisted file and a memory-mapped file
// The size tells us the size of the index and where to write the next entry appended to the index.
type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

// newIndex creates an index for the given file. We create te index and
// save the current size of the file so we can track the amount of data in the
// index file as we add index entries.  We grow the file to the max index size before
// memory-mapping the file and the return the created index to the caller.
func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size())
	if err = os.Truncate(
		f.Name(), int64(c.Segment.MaxIndexBytes),
	); err != nil {
		return nil, err
	}
	if idx.mmap, err = gommap.Map(
		idx.file.Fd(),
		gommap.PROT_READ|gommap.PROT_WRITE,
		gommap.MAP_SHARED,
	); err != nil {
		return nil, err
	}

	return idx, nil
}

// Read takes in an offset and returns the associated record's position in
// the store. The given offset is relative to the segement's base offset.
//
// 0 is always the offset of the index's first entry. We use relative offsets to reduce
// the size of the indexes by storing offsets as uint32s.
func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}
	pos = uint64(out) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

// Write appends the given offset and position to the index.
//
// It validated that we have space to write the entry, if there's space
// we encode the offset and postion and write them to the memory-mapped file.
// Then we increment the position where the next write will go.

func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += uint64(entWidth)
	return nil
}

// Name returns the indes's file path
func (i *index) Name() string {
	return i.file.Name()
}

// Close makes sure the memory-mapped file has synced its data to the
// persisted file and that the persisted fie has flushed its contents to stable storage
// Then it truncates the persisted file to the amount of data that's actually in it and closes the file.
func (i *index) Close() error {
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	if err := i.file.Sync(); err != nil {
		return err
	}
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()
}
