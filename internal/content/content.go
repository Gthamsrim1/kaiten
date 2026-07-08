package content

type Content interface {
	Read(offset int64, p []byte) (int, error)
	Write(offset int64, p []byte) (int, error)
	Size() uint64
	Resize(size uint64) error

	Bytes() []byte
}