package packet

import (
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
)

// Protocol Level
type Version3 byte

const (
	Version311 Version3 = 4
	Version31  Version3 = 3
)

func (v Version3) String() string {
	if v == Version31 {
		return "MQIsdp"
	} else if v == Version311 {
		return "MQTT"
	}
	return "Unknown"
}

type ConnectPacket struct {
	ClientID string
	Username string
	Password string

	CleanSession bool
	Version      Version3
	KeepAlive    uint16

	Will *Message
}

func NewConnectPacket() *ConnectPacket {
	return &ConnectPacket{
		CleanSession: true,
		Version:      Version311,
	}
}

func (p *ConnectPacket) Type() Type {
	return CONNECT
}

func (p *ConnectPacket) Decode(src []byte) (int, error) {
	index := 0

	// decode header
	n, _, _, err := headerDecode(src[index:], CONNECT)
	index += n
	if err != nil {
		return index, err
	}

	// decode variable header
	// read protocol name
	vName, n, err := readString(src[index:])
	index += n
	if err != nil {
		return index, Error(p.Type().String(), "read protocol name", nil, err)
	}
	// check protocol version string
	if vName != Version311.String() && vName != Version31.String() {
		return index, Error(p.Type().String(), "version name byte",
			string(Version311)+"or"+string(Version31), vName)
	}

	// read protocol level
	if len(src[index:]) < 1 {
		return index, Error(p.Type().String(), "buffer size", 1, len(src[index:]))
	}
	protocolL := Version3(src[index])
	index += 1
	if protocolL != Version31 && protocolL != Version311 {
		return index, Error(p.Type().String(), "protocol level", Version31, protocolL)
	}
	p.Version = protocolL

	// read connect packet flags
	if len(src[index:]) < 1 {
		return index, Error(p.Type().String(), "buffer size", 1, len(src[index:]))
	}
	flags := src[index]
	index += 1

	usernameFlag := (flags >> 7) & 0x01
	passwordFlag := (flags >> 6) & 0x01
	willRetainFlag := (flags >> 5) & 0x01
	willQosFlag := Qos((flags >> 4) & 0x03)
	willFlag := (flags >> 2) & 0x01
	cleanSessionFlag := (flags >> 1) & 0x01
	reservedFlag := flags & 0x01

	if reservedFlag != 0 {
		return index, Error(p.Type().String(), "reserved flag", 0, reservedFlag)
	}

	if usernameFlag == 1 && passwordFlag != 1 {
		return index, Error(p.Type().String(), "password flag", 1, passwordFlag)
	}

	if !willQosFlag.Valid() {
		return index, Error(p.Type().String(), "will Qos Flag", "0 1 2", willQosFlag)
	}

	if willFlag == 0 && (willQosFlag != 0 || willRetainFlag != 0) {
		return index, Error(p.Type().String(), "will Qos Flag or will Retain Flag", 0, willQosFlag)
	}

	if willFlag == 1 {
		p.Will = &Message{
			QOS:    willQosFlag,
			Retain: willRetainFlag == 1,
		}
	}
	p.CleanSession = cleanSessionFlag == 1

	// read keepalive
	if len(src[index:]) < 2 {
		return index, Error(p.Type().String(), "buffer length too less", 2, len(src[index:]))
	}
	p.KeepAlive = binary.BigEndian.Uint16(src[index:])
	index += 2

	// decode payload
	// read ClientID
	p.ClientID, n, err = readString(src[index:])
	index += n
	if err != nil {
		return index, Error(p.Type().String(), "client id ", nil, err)
	}

	if willFlag == 1 {
		p.Will.Topic, n, err = readString(src[index:])
		index += n
		if err != nil {
			return index, Error(p.Type().String(), "read will Topic ", nil, err)
		}

		p.Will.Payload, n, err = readBytes(src[index:])
		index += n
		if err != nil {
			return index, Error(p.Type().String(), "read will Payload ", nil, err)
		}
	}
	if usernameFlag == 1 {
		p.Username, n, err = readString(src[index:])
		index += n
		if err != nil {
			return index, Error(p.Type().String(), "read Username", nil, err)
		}
		p.Password, n, err = readString(src[index:])
		index += n
		if err != nil {
			return index, Error(p.Type().String(), "read Password", nil, err)
		}
	}
	return index, nil
}

func (p *ConnectPacket) Encode(dst []byte) (int, error) {
	total := 0

	// encode header
	n, err := headerEncode(dst[total:], 0, p.len(), p.Len(), CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// set default version byte
	if p.Version == 0 {
		p.Version = Version311
	}

	// check version byte
	if p.Version != Version311 && p.Version != Version31 {
		return total, Error(string(CONNECT), " protocol version", "", p.Version)
	}

	// write version string, length has been checked beforehand
	if p.Version == Version311 {
		n, _ = writeBytes(dst[total:], []byte(string(Version311)), CONNECT)
		total += n
	} else if p.Version == Version31 {
		n, _ = writeBytes(dst[total:], []byte(string(Version31)), CONNECT)
		total += n
	}

	// write version value
	dst[total] = byte(p.Version)
	total++

	var connectFlags byte

	// set username flag
	if len(p.Username) > 0 {
		connectFlags |= 128 // 10000000
	} else {
		connectFlags &= 127 // 01111111
	}

	// set password flag
	if len(p.Password) > 0 {
		connectFlags |= 64 // 01000000
	} else {
		connectFlags &= 191 // 10111111
	}

	// set will flag
	if p.Will != nil {
		connectFlags |= 0x4 // 00000100

		// check will topic length
		if len(p.Will.Topic) == 0 {
			return total, fmt.Errorf("[%s] will topic is empty", p.Type())
		}

		// check will qos
		if !p.Will.QOS.Valid() {
			return total, fmt.Errorf("[%s] invalid will qos level %d", p.Type(), p.Will.QOS)
		}

		// set will qos flag
		connectFlags = (connectFlags & 231) | (byte(p.Will.QOS) << 3) // 231 = 11100111

		// set will retain flag
		if p.Will.Retain {
			connectFlags |= 32 // 00100000
		} else {
			connectFlags &= 223 // 11011111
		}

	} else {
		connectFlags &= 251 // 11111011
	}

	// check client id and clean session
	if len(p.ClientID) == 0 && !p.CleanSession {
		return total, fmt.Errorf("[%s] clean session must be 1 if client id is zero length", p.Type())
	}

	// set clean session flag
	if p.CleanSession {
		connectFlags |= 0x2 // 00000010
	} else {
		connectFlags &= 253 // 11111101
	}

	// write connect flags
	dst[total] = connectFlags
	total++

	// write keep alive
	binary.BigEndian.PutUint16(dst[total:], p.KeepAlive)
	total += 2

	// write client id
	n, err = writeString(dst[total:], p.ClientID, p.Type())
	total += n
	if err != nil {
		return total, err
	}

	// write will topic and payload
	if p.Will != nil {
		n, err = writeString(dst[total:], p.Will.Topic, p.Type())
		total += n
		if err != nil {
			return total, err
		}

		n, err = writeBytes(dst[total:], p.Will.Payload, p.Type())
		total += n
		if err != nil {
			return total, err
		}
	}

	if len(p.Username) == 0 && len(p.Password) > 0 {
		return total, fmt.Errorf("[%s] password set without username", p.Type())
	}

	// write username
	if len(p.Username) > 0 {
		n, err = writeString(dst[total:], p.Username, p.Type())
		total += n
		if err != nil {
			return total, err
		}
	}

	// write password
	if len(p.Password) > 0 {
		n, err = writeString(dst[total:], p.Password, p.Type())
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
	return 0, nil
}

func (p *ConnectPacket) len() int {
	remainLength := 0
	// version name
	remainLength += len(p.Version.String()) + 2
	// protcol level
	remainLength++
	// flags
	remainLength++
	// keepalive
	remainLength += 2
	// client id
	remainLength += len(p.ClientID) + 2

	// will
	if p.Will != nil {
		remainLength += len(p.Will.Topic) + len(p.Will.Payload) + 4
	}
	// username password
	if p.Username != "" {
		remainLength += len(p.Username) + len(p.Password) + 4
	}
	return remainLength
}

func (p *ConnectPacket) Len() int {
	l := p.len()
	return l + headerLen(l)
}

func readString(buf []byte) (string, int, error) {
	index := 0
	if len(buf) < 2 {
		return "", index, errors.New("buf length less than 2")
	}

	l := int(binary.BigEndian.Uint16(buf[index:]))
	index += 2

	if len(buf[index:]) < l {
		return "", index, errors.New("buf length error")
	}

	return string(buf[index : index+l]), index + l, nil
}

func readBytes(buf []byte) ([]byte, int, error) {
	index := 0
	if len(buf) < 2 {
		return nil, index, errors.New("buf length less than 2")
	}

	l := int(binary.BigEndian.Uint16(buf[index:]))
	index += 2

	if len(buf[index:]) < l {
		return nil, index, errors.New("buf length error")
	}
	r := make([]byte, l)
	copy(r, buf[index:index+l])
	return r, index + l, nil
}

const maxLPLength uint16 = 65535

func writeString(dst []byte, s string, t Type) (int, error) {
	return writeBytes(dst, []byte(s), t)
}

func writeBytes(dst []byte, b []byte, t Type) (int, error) {
	total, n := 0, len(b)

	if n > int(maxLPLength) {
		return 0, Error(string(t), "string length ", "less than 65535 byte", n)
	}

	if len(dst) < 2+n {
		return 0, Error(string(t), "insufficient buffer size ", 2+n, len(dst))
	}

	binary.BigEndian.PutUint16(dst, uint16(n))
	total += 2

	copy(dst[total:], b)
	total += n

	return total, nil
}
