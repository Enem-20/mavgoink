package main

import (
	"log"
	"mavgoink/system"
)

func main() {
	sys := system.NewSystem(system.MAVLINK_VERSION_2, 255, "Default")

	msg := sys.CreateMessage(1, 0, 9)
	msg.PushBytes([]byte{0xFD, 0xFD, 0xFD})
	msg.PushUint32(0xDEADBEEF)
	msg.PushByte(0xFD)
	msg.PushByte(0xFD)
	log.Printf("msg: %#X\n", string(msg.GetRawMessage()))
}
