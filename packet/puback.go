package packet

import (
	"encoding/binary"
	"fmt"
)

// A PubackPacket is the response to a PublishPacket with QOS level 1.
type PubackPacket struct {
	// The packet identifier.
	ID uint16
}

// NewPubackPacket creates a new PubackPacket.
func NewPubackPacket() *PubackPacket {
	return &PubackPacket{}
}

// Type returns the packets type.
func (p *PubackPacket) Type() Type {
	return PUBACK
}

// Len returns the byte length of the encoded packet.
func (p *PubackPacket) Len() int {
	return headerLen(2) + 2
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *PubackPacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src, PUBACK)
	total += hl
	if err != nil {
		return total, err
	}

	// check remaining length
	if rl != 2 {
		return total, fmt.Errorf("[%s] expected remaining length to be 2", PUBACK)
	}

	// read packet id
	packetID := binary.BigEndian.Uint16(src[total:])
	total += 2

	// check packet id
	if packetID == 0 {
		return total, fmt.Errorf("[%s] packet id must be grater than zero", PUBACK)
	}

	p.ID = packetID
	return total, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *PubackPacket) Encode(dst []byte) (int, error) {
	total := 0

	// check packet id
	if p.ID == 0 {
		return total, fmt.Errorf("[%s] packet id must be grater than zero", PUBACK)
	}

	// encode header
	n, err := headerEncode(dst[total:], 0, 2, p.Len(), PUBACK)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	binary.BigEndian.PutUint16(dst[total:], p.ID)
	total += 2

	return total, nil
}

// String returns a string representation of the packet.
func (p *PubackPacket) String() string {
	return fmt.Sprintf("<PubackPacket ID=%d>", p.ID)
}
