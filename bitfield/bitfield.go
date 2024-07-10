package bitfield

type Bitfield []byte

func (bf *Bitfield) HasPiece(index int) bool {
	byteNumber := index / 8
	byteOffset := index % 8
	if byteNumber < 0 || byteNumber >= len(*bf) {
		return false
	}
	return (*bf)[byteNumber]>>uint(7-byteOffset)&1 != 0
}

func (bf *Bitfield) SetPiece(index int) {
	byteNumber := index / 8
	byteOffset := index % 8
	if byteNumber < 0 || len(*bf) >= byteNumber {
		return
	}
	(*bf)[byteNumber] |= 1 << uint(7-byteOffset)
}
