package packet

type DisConnectPacket struct {
}

func NewDisconnectPacket() *DisConnectPacket {
	return &DisConnectPacket{}
}

func (p *DisConnectPacket) Type() Type {
	return DISCONNECT
}

func (p *DisConnectPacket) Decode(src []byte) (int, error) {
	// decode header
	hl, _, rl, err := headerDecode(src, DISCONNECT)

	// check remaining length
	if rl != 0 {
		return hl, Error(string(DISCONNECT), "expected zero remaining length", "", "")
	}

	return hl, err
}

func (p *DisConnectPacket) Encode(dst []byte) (int, error) {
	return headerEncode(dst, 0, 0, p.Len(), DISCONNECT)
}

func (p *DisConnectPacket) Len() int {
	return headerLen(0)
}
