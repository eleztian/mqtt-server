package packet

import "errors"

//go:generate stringer -type=Type

// Type represents the MQTT packet types.
type Type byte

// All packet types.
const (
	_           Type = iota // 保留
	CONNECT                 // 请求连接
	CONNACK                 // 请求应答
	
	PUBLISH                 // 发布消息
	PUBACK                  // 发布应答
	PUBREC                  // 发布已接收，保证传递1
	PUBREL                  // 发布释放，保证传递2
	PUBCOMP                 // 发布完成，保证传递3
	SUBSCRIBE               // 订阅请求
	SUBACK                  // 订阅应答
	UNSUBSCRIBE             // 取消订阅
	UNSUBACK                // 取消订阅应答

	PINGREQ                 // ping请求
	PINGRESP                // ping响应
	DISCONNECT              // 断开连接
	_                       // 保留
)

// DefaultFlags returns the default flag values for the packet type, as defined
// by the MQTT spec, except for PUBLISH.
func (t Type) defaultFlags() byte {
	switch t {
	case CONNECT:
		return 0
	case CONNACK:
		return 0
	case PUBACK:
		return 0
	case PUBREC:
		return 0
	case PUBREL:
		return 2 // 00000010
	case PUBCOMP:
		return 0
	case SUBSCRIBE:
		return 2 // 00000010
	case SUBACK:
		return 0
	case UNSUBSCRIBE:
		return 2 // 00000010
	case UNSUBACK:
		return 0
	case PINGREQ:
		return 0
	case PINGRESP:
		return 0
	case DISCONNECT:
		return 0
	}

	return 0
}

func (t Type) New() (Packet, error) {
	switch t {
	case CONNECT:
		return NewConnectPacket(), nil
	}
	return nil, errors.New("unknown type")
}

// QOS

type Qos byte

const (
	// QOSAtMostOnce 00表示最多一次
	QOSAtMostOnce Qos = iota

	// QOSAtLeastOnce 01表示至少一次
	QOSAtLeastOnce

	// QOSExactlyOnce 10表示一次
	QOSExactlyOnce

	// QOSFailure indicates that there has been an error while subscribing
	// to a specific topic.
	QOSFailure = 0x80
)

func validQos(q byte) bool {
	if q > 2 || q < 0 {
		return false
	}
	return true
}