package packet

import (
	"encoding/binary"
	"fmt"
)

// A PublishPacket is sent from a client to a server or from server to a client
// to transport an application message.
type PublishPacket struct {
	// The message to publish.
	Message Message

	// If the Dup flag is set to false, it indicates that this is the first
	// occasion that the client or server has attempted to send this
	// PublishPacket. If the dup flag is set to true, it indicates that this
	// might be re-delivery of an earlier attempt to send the packet.
	Dup bool

	// The packet identifier.
	ID uint16
}

// NewPublishPacket creates a new PublishPacket.
func NewPublishPacket() *PublishPacket {
	return &PublishPacket{}
}

// Type returns the packets type.
func (p *PublishPacket) Type() Type {
	return PUBLISH
}

// String returns a string representation of the packet.
func (p *PublishPacket) String() string {
	return fmt.Sprintf("<PublishPacket ID=%d Message=%s Dup=%t>",
		p.ID, p.Message.String(), p.Dup)
}

// Len returns the byte length of the encoded packet.
func (p *PublishPacket) Len() int {
	ml := p.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *PublishPacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, flags, rl, err := headerDecode(src[total:], PUBLISH)
	total += hl
	if err != nil {
		return total, err
	}

	// read flags
	p.Dup = ((flags >> 3) & 0x1) == 1
	p.Message.Retain = (flags & 0x1) == 1
	p.Message.QOS = Qos((flags >> 1) & 0x3)

	// check qos
	if !p.Message.QOS.Valid() {
		return total, fmt.Errorf("[%s] invalid QOS level (%d)", p.Type(), p.Message.QOS)
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", p.Type(), total+2, len(src))
	}

	n := 0

	// read topic
	p.Message.Topic, n, err = readString(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	if p.Message.QOS != 0 {
		// check buffer length
		if len(src) < total+2 {
			return total, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", p.Type(), total+2, len(src))
		}

		// read packet id
		p.ID = binary.BigEndian.Uint16(src[total:])
		total += 2

		// check packet id
		if p.ID == 0 {
			return total, fmt.Errorf("[%s] packet id must be grater than zero", p.Type())
		}
	}

	// calculate payload length
	l := int(rl) - (total - hl)

	// read payload
	if l > 0 {
		p.Message.Payload = make([]byte, l)
		copy(p.Message.Payload, src[total:total+l])
		total += len(p.Message.Payload)
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *PublishPacket) Encode(dst []byte) (int, error) {
	total := 0

	// check topic length
	if len(p.Message.Topic) == 0 {
		return total, fmt.Errorf("[%s] topic name is empty", p.Type())
	}

	flags := byte(0)

	// set dup flag
	if p.Dup {
		flags |= 0x8 // 00001000
	} else {
		flags &= 247 // 11110111
	}

	// set retain flag
	if p.Message.Retain {
		flags |= 0x1 // 00000001
	} else {
		flags &= 254 // 11111110
	}

	// check qos
	if !p.Message.QOS.Valid() {
		return 0, fmt.Errorf("[%s] invalid QOS level %d", p.Type(), p.Message.QOS)
	}

	// check packet id
	if p.Message.QOS > 0 && p.ID == 0 {
		return total, fmt.Errorf("[%s] packet id must be grater than zero", p.Type())
	}

	// set qos
	flags = (flags & 249) | (byte(p.Message.QOS) << 1) // 249 = 11111001

	// encode header
	n, err := headerEncode(dst[total:], flags, p.len(), p.Len(), PUBLISH)
	total += n
	if err != nil {
		return total, err
	}

	// write topic
	n, err = writeString(dst[total:], p.Message.Topic, p.Type())
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	if p.Message.QOS != 0 {
		binary.BigEndian.PutUint16(dst[total:], uint16(p.ID))
		total += 2
	}

	// write payload
	copy(dst[total:], p.Message.Payload)
	total += len(p.Message.Payload)

	return total, nil
}

// Returns the payload length.
func (p *PublishPacket) len() int {
	total := 2 + len(p.Message.Topic) + len(p.Message.Payload)
	if p.Message.QOS != 0 {
		total += 2
	}

	return total
}
