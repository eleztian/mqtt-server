package packet

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// A Subscription is a single subscription in a SubscribePacket.
type Subscription struct {
	// The topic to subscribe.
	Topic string

	// The requested maximum QOS level.
	QOS Qos
}

func (s *Subscription) String() string {
	return fmt.Sprintf("%q=>%d", s.Topic, s.QOS)
}

// A SubscribePacket is sent from the client to the server to create one or
// more Subscriptions. The server will forward application messages that match
// these subscriptions using PublishPackets.
type SubscribePacket struct {
	// The subscriptions.
	Subscriptions []Subscription

	// The packet identifier.
	ID uint16
}

// NewSubscribePacket creates a new SUBSCRIBE packet.
func NewSubscribePacket() *SubscribePacket {
	return &SubscribePacket{}
}

// Type returns the packets type.
func (p *SubscribePacket) Type() Type {
	return SUBSCRIBE
}

// String returns a string representation of the packet.
func (p *SubscribePacket) String() string {
	var subscriptions []string

	for _, t := range p.Subscriptions {
		subscriptions = append(subscriptions, t.String())
	}

	return fmt.Sprintf("<SubscribePacket ID=%d Subscriptions=[%s]>",
		p.ID, strings.Join(subscriptions, ", "))
}

// Len returns the byte length of the encoded packet.
func (p *SubscribePacket) Len() int {
	ml := p.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *SubscribePacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src[total:], SUBSCRIBE)
	total += hl
	if err != nil {
		return total, err
	}

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

	// reset subscriptions
	p.Subscriptions = p.Subscriptions[:0]

	// calculate number of subscriptions
	sl := int(rl) - 2

	for sl > 0 {
		// read topic
		t, n, err := readString(src[total:])
		total += n
		if err != nil {
			return total, err
		}

		// check buffer length
		if len(src) < total+1 {
			return total, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", p.Type(), total+1, len(src))
		}

		// read qos and add subscription
		p.Subscriptions = append(p.Subscriptions, Subscription{t, Qos(src[total])})
		total++

		// decrement counter
		sl = sl - n - 1
	}

	// check for empty subscription list
	if len(p.Subscriptions) == 0 {
		return total, fmt.Errorf("[%s] empty subscription list", p.Type())
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *SubscribePacket) Encode(dst []byte) (int, error) {
	total := 0

	// check packet id
	if p.ID == 0 {
		return total, fmt.Errorf("[%s] packet id must be grater than zero", p.Type())
	}

	// encode header
	n, err := headerEncode(dst[total:], 0, p.len(), p.Len(), SUBSCRIBE)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	binary.BigEndian.PutUint16(dst[total:], uint16(p.ID))
	total += 2

	for _, t := range p.Subscriptions {
		// write topic
		n, err := writeString(dst[total:], t.Topic, p.Type())
		total += n
		if err != nil {
			return total, err
		}

		// write qos
		dst[total] = byte(t.QOS)

		total++
	}

	return total, nil
}

// Returns the payload length.
func (p *SubscribePacket) len() int {
	// packet ID
	total := 2

	for _, t := range p.Subscriptions {
		total += 2 + len(t.Topic) + 1
	}

	return total
}
