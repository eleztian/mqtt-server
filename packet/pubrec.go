package packet

import (
	"encoding/binary"
	"fmt"
)

// A PubackPacket is the response to a PublishPacket with QOS level 1.
type PubrecPacket struct {
	// The packet identifier.
	ID uint16
}

// NewPubackPacket creates a new PubackPacket.
func NewPubrecPacket() *PubrecPacket {
	return &PubrecPacket{}
}

// Type returns the packets type.
func (p *PubrecPacket) Type() Type {
	return PUBREC
}

// Len returns the byte length of the encoded packet.
func (p *PubrecPacket) Len() int {
	return headerLen(2) + 2
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *PubrecPacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src, PUBREC)
	total += hl
	if err != nil {
		return total, err
	}

	// check remaining length
	if rl != 2 {
		return total, fmt.Errorf("[%s] expected remaining length to be 2", PUBREC)
	}

	// read packet id
	packetID := binary.BigEndian.Uint16(src[total:])
	total += 2

	// check packet id
	if packetID == 0 {
		return total, fmt.Errorf("[%s] packet id must be grater than zero", PUBREC)
	}

	p.ID = packetID
	return total, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *PubrecPacket) Encode(dst []byte) (int, error) {
	total := 0

	// check packet id
	if p.ID == 0 {
		return total, fmt.Errorf("[%s] packet id must be grater than zero", PUBREC)
	}

	// encode header
	n, err := headerEncode(dst[total:], 0, 2, p.Len(), PUBREC)
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
func (p *PubrecPacket) String() string {
	return fmt.Sprintf("<PubrecPacket ID=%d>", p.ID)
}
