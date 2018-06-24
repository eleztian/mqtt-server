package packet

type ConnackCode byte

const (
	ConnectionAccepted ConnackCode = iota
	ErrInvalidProtocolVersion
	ErrIdentifierRejected
	ErrServerUnavailable
	ErrBadUsernameOrPassword
	ErrNotAuthorized
)
// Valid checks if the ConnackCode is valid.
func (cc ConnackCode) Valid() bool {
	return cc <= 5
}

// Error returns the corresponding error string for the ConnackCode.
func (cc ConnackCode) Error() string {
	switch cc {
	case ConnectionAccepted:
		return "connection accepted"
	case ErrInvalidProtocolVersion:
		return "connection refused: unacceptable protocol version"
	case ErrIdentifierRejected:
		return "connection refused: identifier rejected"
	case ErrServerUnavailable:
		return "connection refused: server unavailable"
	case ErrBadUsernameOrPassword:
		return "connection refused: bad user name or password"
	case ErrNotAuthorized:
		return "connection refused: not authorized"
	}

	return "unknown error"
}

type ConnectAckPacket struct {
	// 如果服务端收到一个CleanSession为0的连接，
	// 当前会话标志的值取决于服务端是否已经保存了ClientId对应客户端的会话状态。
	// 如果服务端已经保存了会话状态，它必须将CONNACK报文中的当前会话标志设置为1。
	// 如果服务端没有已保存的会话状态，它必须将CONNACK报文中的当前会话设置为0。
	// 还需要将CONNACK报文中的返回码设置为0
	SessionPresent bool
	// 如果服务端发送了一个包含非零返回码的CONNACK报文，那么它必须关闭网络连接
	ReturnCode     ConnackCode
}

func NewConnackPacket() *ConnectAckPacket {
	return &ConnectAckPacket{}
}

func (p *ConnectAckPacket) Type() Type {
	return CONNACK
}

func (p *ConnectAckPacket) Len() int {
	return headerLen(2) + 2
}

func (p *ConnectAckPacket) Encode(dst []byte) (int, error) {
	index := 0
	l, err := headerEncode(dst[index:], 0, 2, p.Len(), CONNACK)
	index += l
	if err != nil {
		return index, err
	}

	// set session present flag
	if p.SessionPresent {
		dst[index] = 1 // 00000001
	} else {
		dst[index] = 0 // 00000000
	}
	index++

	// check return code
	if !p.ReturnCode.Valid() {
		return index, Error(string(CONNACK), "invalid return code", "<=5", p.ReturnCode)
	}

	// set return code
	dst[index] = byte(p.ReturnCode)
	index++
	return index, nil
}

func (p *ConnectAckPacket) Decode(src []byte) (int, error) {
	index := 0

	l, _, _, err := headerDecode(src[index:], CONNACK)
	index += l
	if err != nil {
		return index, err
	}
	// read connet ack flag
	if len(src) < index+1 {
		return index, Error(string(CONNACK), "buffer size", index+1, len(src))
	}
	connetOkFlag := src[index]
	index++
	if connetOkFlag|0x01 != 1 {
		return index, Error(string(CONNACK), "connect ack flag", 0, connetOkFlag)
	}
	p.SessionPresent = (connetOkFlag & 0x01) == 1

	// read connect ack return code
	if len(src) < index+1 {
		return index, Error(string(CONNACK), "buffer size", index+1, len(src))
	}
	p.ReturnCode = ConnackCode(src[index])
	index++

	// check return code
	if !p.ReturnCode.Valid() {
		return index, Error(string(CONNACK), "invalid return code", "<=5", p.ReturnCode)
	}

	return index, nil
}
