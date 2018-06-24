package packet

type PingReqPacket struct {
}

func NewPingReqPacket() *PingReqPacket {
	return &PingReqPacket{}
}

func (p *PingReqPacket) Type() Type {
	return PINGREQ
}

func (p *PingReqPacket) Decode(src []byte) (int, error) {
	// decode header
	hl, _, rl, err := headerDecode(src, PINGREQ)

	// check remaining length
	if rl != 0 {
		return hl, Error(string(PINGREQ), "expected zero remaining length", "", "")
	}

	return hl, err
}

func (p *PingReqPacket) Encode(dst []byte) (int, error) {
	return headerEncode(dst, 0, 0, p.Len(), PINGREQ)
}

func (p *PingReqPacket) Len() int {
	return headerLen(0)
}