package packet

// A Message bundles data that is published between brokers and clients.
type Message struct {
	// The Topic of the message.
	Topic string

	// The Payload of the message.
	Payload []byte

	// The QOS indicates the level of assurance for delivery.
	QOS Qos

	// If the Retain flag is set to true, the server must store the message,
	// so that it can be delivered to future subscribers whose subscriptions
	// match its topic name.
	Retain bool
}

func (m *Message) String() string {
	return string(m.Payload)
}