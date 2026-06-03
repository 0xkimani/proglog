package log

// Config wraps up the code for store and index types, which make up the
// lowest level of our log. It configures the max size of a segement' store and index.
type Config struct {
	Segment struct {
		MaxStoreBytes uint64
		MaxIndexBytes uint64
		InitialOffset uint64
	}
}
