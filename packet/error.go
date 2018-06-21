package packet

import "fmt"

//go:generate stringer -type=PacketError

// Packet error
type PacketError int

const (
	_ PacketError = iota
	BUFFER_SIZE_ERROR
	TYPE_ERROR
	FLAGS_ERROR
	REMAIN_LENGTH_ERROR
	REMAIN_BUFFERR_SIZE_ERROR
)

func (i PacketError) Error() string {
	return fmt.Sprintf("%d:%s", i, i)
}
