package packet

import (
	"encoding/binary"
	"errors"
)

type Packet interface {
	Type() Type
	Len() int
	Decode(src []byte) (int, error)
	Encode(dst []byte) (int, error)
}

func NewPacket(t Type) (Packet, error) {
	switch t {
	case CONNECT:
		return &ConnectPacket{}, nil
	case CONNACK:
		return &ConnectAckPacket{}, nil
	}
	return nil, errors.New("unknown packet type")
}


// 解析前两个字节,尝试获取下一个packet的类型和长度
func DetectPacket(src []byte) (int, Type) {
	// check for minimum size
	if len(src) < 2 {
		return 0, 0
	}

	// get type
	t := Type(src[0] >> 4)

	// get remaining length
	rl, n := binary.Uvarint(src[1:])
	if n <= 0 {
		return 0, 0
	}

	return 1 + n + int(rl), t
}