package packet

import (
	"testing"
	"fmt"
)

func TestDecode(t *testing.T) {
	cp := ConnectPacket{}
	_ ,err := cp.Decode([]byte{1})
	if err != nil {
		fmt.Printf("%+v", err)
	}
}