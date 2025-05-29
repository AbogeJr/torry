package bitfield

type Bitfield []byte

func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	bitIndex := index % 8

	mask := bf[byteIndex] >> (7 - bitIndex) & 1

	return mask != 0
}

func (bf Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	bitIndex := index % 8

	bf[byteIndex] |= 1 << (7 - bitIndex)
}
