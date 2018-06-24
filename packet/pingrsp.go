package packet

type PingRspPacket struct {
}

func (p *PingRspPacket) Type() Type {
	return PINGRESP
}

func (p *PingRspPacket) Decode(src []byte) (int, error) {
	// decode header
	hl, _, rl, err := headerDecode(src, PINGRESP)

	// check remaining length
	if rl != 0 {
		return hl, Error(string(PINGRESP), "expected zero remaining length", "", "")
	}

	return hl, err
}

func (p *PingRspPacket) Encode(dst []byte) (int, error) {
	return headerEncode(dst, 0, 0, p.Len(), PINGRESP)
}

func (p *PingRspPacket) Len() int {
	return headerLen(0)
}