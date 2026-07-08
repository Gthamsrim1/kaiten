package chunk

type Chunk struct {
	Offset int64
	Length uint32
	Hash   [32]byte
	Data   []byte
}
