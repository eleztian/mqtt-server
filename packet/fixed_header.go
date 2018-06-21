package packet

import (
	"encoding/binary"
	"github.com/pkg/errors"
)

const maxRemainingLength = 268435455 // 256 MB

func headerDecode(src []byte, t Type) (index int, flags byte, rl int, err error) {
	// check buffer size
	if len(src) < 2 {
		err = errors.Wrap(BUFFER_SIZE_ERROR, "HeaderDecode")
		return
	}

	// read type
	typeAndFlags := src[index]
	decodeType := Type(typeAndFlags >> 4)
	flags = typeAndFlags & 0x0f
	index++

	// check type
	if t != decodeType {
		err = errors.Wrap(TYPE_ERROR, "HeaderDecode")
		return
	}

	// check flags
	if t != PUBLISH && t.defaultFlags() != flags {
		err = errors.Wrap(FLAGS_ERROR, "HeaderDecode")
		return
	}

	// read remaining length
	_rl, m := binary.Uvarint(src[index:])
	rl = int(_rl)

	// check length
	if m <= 0 {
		err = errors.Wrap(REMAIN_LENGTH_ERROR, "HeaderDecode")
		return
	}
	index += m
	if len(src[index:]) != rl {
		err = errors.Wrap(REMAIN_BUFFERR_SIZE_ERROR, "HeaderDecode")
		return
	}
	return
}

func headerEncode(dst []byte, flags byte, rl int, al int, t Type) (index int, err error) {
	// check length
	if len(dst) < al {
		err = errors.Wrap(BUFFER_SIZE_ERROR, "HeaderEncode")
		return
	}

	// check remaining length
	if rl > maxRemainingLength {
		err = errors.Wrap(REMAIN_LENGTH_ERROR, "HeaderDecode")
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
		err = errors.Wrap(REMAIN_BUFFERR_SIZE_ERROR, "HeaderEncode")
		return
	}

	return
}
