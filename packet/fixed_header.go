package packet

import (
	"encoding/binary"
	"fmt"
)

const maxRemainingLength = 268435455 // 256 MB

func headerDecode(src []byte, t Type) (index int, flags byte, rl int, err error) {
	// check buffer size
	if len(src) < 2 {
		err = fmt.Errorf("[%s] insufficient buffer size, want more than %d, got %d", t, 2, len(src))
		return
	}

	// read type
	typeAndFlags := src[index]
	decodeType := Type(typeAndFlags >> 4)
	flags = typeAndFlags & 0x0f
	index++

	// check type
	if t != decodeType {
		err = fmt.Errorf("[%s] deocode type error want %s get %s", t, t, decodeType)
		return
	}

	// check flags
	if t != PUBLISH && t.defaultFlags() != flags {
		err = fmt.Errorf("[%s] flag error want %d get %d", t, t.defaultFlags(), flags)
		return
	}

	// read remaining length
	_rl, m := binary.Uvarint(src[index:])
	rl = int(_rl)

	// check length
	if m <= 0 {
		err = fmt.Errorf("[%s] remaining length error, want %d get %d", t, t.defaultFlags(), flags)
		return
	}
	index += m
	if len(src[index:]) < rl {
		err = fmt.Errorf("[%s] leave buffer length too less, want %d get %d", t, rl, len(src[index:]))
		return
	}
	return
}

func headerEncode(dst []byte, flags byte, rl int, al int, t Type) (index int, err error) {
	// check length
	if len(dst) < al {
		err = fmt.Errorf("[%s] insufficient buffer size, want more than %d, got %d", t, 2, len(dst))
		return
	}

	// check remaining length
	if rl > maxRemainingLength {
		err = fmt.Errorf("[%s]  remaining length should less than %d, but got %d", t, maxRemainingLength, rl)
		return
	}

	// create type and flags
	typeAndFlags := byte(t)<<4 | (t.defaultFlags() & 0x0f)
	typeAndFlags |= flags
	dst[index] = typeAndFlags
	index++

	// write remaining length
	n := binary.PutUvarint(dst[index:], uint64(rl))
	index += n

	// check buf length
	if len(dst) < (1 + n + rl) {
		err = fmt.Errorf("[%s]  buffer size error, want %d got %d", t, 1+n+rl, len(dst))
		return
	}

	return
}
func headerLen(rl int) int {
	// packet type and flag byte
	total := 1

	if rl <= 127 {
		total++
	} else if rl <= 16383 {
		total += 2
	} else if rl <= 2097151 {
		total += 3
	} else {
		total += 4
	}

	return total
}