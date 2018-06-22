package packet

import (
	"encoding/binary"
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
	willQosFlag := (flags >> 4) & 0x03
	willFlag := (flags >> 2) & 0x01
	cleanSessionFlag := (flags >> 1) & 0x01
	reservedFlag := flags & 0x01

	if reservedFlag != 0 {
		return index, Error(p.Type().String(), "reserved flag", 0, reservedFlag)
	}

	if usernameFlag == 1 && passwordFlag != 1 {
		return index, Error(p.Type().String(), "password flag", 1, passwordFlag)
	}

	if !validQos(willQosFlag) {
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

	return 0, nil
}

func (p *ConnectPacket) Len() int {
	remainLength := 0
	// version name
	remainLength += len(p.Version.String()) + 2
	// protcol level
	remainLength ++
	// flags
	remainLength ++
	// keepalive
	remainLength +=2
	// client id
	remainLength += len(p.ClientID) + 2

	// will
	if p.Will != nil {
		remainLength += len(p.Will.Topic)+ len(p.Will.Payload) + 4
	}
	// username password
	if p.Username != "" {
		remainLength += len(p.Username) + len(p.Password) + 4
	}
	return remainLength + headerLen(remainLength)
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
