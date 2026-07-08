package store

type ChunkRef struct {
	Hash   [32]byte
	Length uint32
}
