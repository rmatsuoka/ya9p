package plan9

import (
	"errors"
	"fmt"
)

var (
	errStringTooLong = errors.New("string too long")
	errInvalidType   = errors.New("invalid type")
)

func Pack(elems ...interface{}) ([]byte, error) {
	p := make([]byte, 0, 128)
	o := 0
	for _, e := range elems {
		if cap(p) < o + 8 {
			np := make([]byte, o, 2*cap(p)+100)
			copy(p, np)
			p = np
		}
		switch e := e.(type) {
		case uint8:
			p = p[:o+1]
			p[o] = byte(e)
			o += 1
		case uint16:
			p = p[:o+2]
			p[o] = byte(e)
			p[o+1] = byte(e >> 8)
			o += 2
		case uint32:
			p = p[:o+4]
			p[o] = byte(e)
			p[o+1] = byte(e >> 8)
			p[o+2] = byte(e >> 16)
			p[o+3] = byte(e >> 24)
			o += 4
		case uint64:
			p = p[:o+8]
			p[o+0] = byte(e >> 0)
			p[o+1] = byte(e >> 8)
			p[o+2] = byte(e >> 16)
			p[o+3] = byte(e >> 24)
			p[o+4] = byte(e >> 32)
			p[o+5] = byte(e >> 40)
			p[o+6] = byte(e >> 48)
			p[o+7] = byte(e >> 56)
			o += 8
		case []byte:
			p = append(p, e...)
			o += len(e)
		case string:
			size := len(e)
			if size >= (1 << 16) {
				return p, errStringTooLong
			}
			p = p[:o+2]
			p[o] = byte(size)
			p[o+1] = byte(size >> 8)
			p = append(p, []byte(e)...)
			o += 2 + size
		default:
			return p, errInvalidType
		}
	}
	return p, nil
}

func MustPack(e ...interface{}) []byte {
	p, err := Pack(e...)
	if err != nil {
		panic(err)
	}
	return p
}

func Unpack(p []byte, elem ...interface{}) (o int, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()
	o = 0
loop:
	for _, e := range elem {
		switch e := e.(type) {
		case *uint8:
			*e = uint8(p[o])
			o += 1
		case *uint16:
			*e = uint16(p[o]) | uint16(p[o+1])<<8
			o += 2
		case *uint32:
			*e = uint32(p[o]) | uint32(p[o+1])<<8 | uint32(p[o+2])<<16 | uint32(p[o+3])<<24
			o += 4
		case *uint64:
			lo := uint32(p[o]) | uint32(p[o+1])<<8 | uint32(p[o+2])<<16 | uint32(p[o+3])<<24
			o += 4
			hi := uint32(p[o]) | uint32(p[o+1])<<8 | uint32(p[o+2])<<16 | uint32(p[o+3])<<24
			o += 4
			*e = uint64(hi)<<32 | uint64(lo)
		case *[]byte:
			*e = p[o:]
			break loop
		case *string:
			size := int(uint16(p[o]) | uint16(p[o+1])<<8)
			o += 2
			*e = string(p[o : o+size])
			o += size
		default:
			return o, errInvalidType
		}
	}
	return o, nil
}
