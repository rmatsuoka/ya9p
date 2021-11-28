package plan9

import (
	"testing"
)

func TestPack(t *testing.T) {
	var u8  uint8  = 8
	var u16 uint16 = 16
	var u32 uint32 = 32
	var u64 uint64 = 64
	t.Log(pack(u8, u16, u32, u64))
}

func TestUnpack(t *testing.T) {
	var u8, uu8  uint8
	var u16, uu16 uint16
	var u32, uu32 uint32
	var u64, uu64 uint64
	u8, u16, u32, u64 = 8, 16, 32, 64
	b := mustPack(u8, u16, u32, u64)
	size, err := unpack(b, &uu8, &uu16, &uu32, &uu64)
	t.Log(size, err, uu8, uu16, uu32, uu64)
}