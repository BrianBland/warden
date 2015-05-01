package warden

import (
	"encoding/binary"
	"syscall"
	"unsafe"
)

func parseDimensions(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

type windowSize struct {
	height, width uint16
	x, y          uint16 // unused
}

func setWindowSize(fd uintptr, width, height uint32) {
	ws := &windowSize{
		width:  uint16(width),
		height: uint16(height),
	}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
