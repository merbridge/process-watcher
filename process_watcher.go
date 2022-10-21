package processwatcher

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type cb_id struct {
	idx uint32
	val uint32
}

type cn_msg struct {
	id cb_id

	seq uint32
	ack uint32

	len   uint16
	flags uint16
	data  [1]uint8
}

const (
	CN_IDX_PROC = 1
	CN_VAL_PROC = 1
)

type Watcher interface {
	Start() error
	Stop()
	Events() <-chan WatchEvent
}

type processWatcher struct {
	sock   int
	events chan WatchEvent
	stop   chan struct{}
}

func NewProcessWatcher() Watcher {
	return &processWatcher{
		events: make(chan WatchEvent),
		stop:   make(chan struct{}),
	}
}

func iovecs2bytes(iovecs []unix.Iovec) [][]byte {
	bs := make([][]byte, len(iovecs))
	for i, v := range iovecs {
		b := make([]byte, v.Len)
		for j := 0; j < int(v.Len); j++ {
			b[j] = *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(v.Base)) + uintptr(j)))
		}
		bs[i] = b
	}
	return bs
}

func (pw *processWatcher) Stop() {
	syscall.Close(pw.sock)
	close(pw.stop)
	close(pw.events)
}

func (pw *processWatcher) Events() <-chan WatchEvent {
	return pw.events
}

func (pw *processWatcher) sendEvents(e WatchEvent) {
	select {
	case <-pw.stop:
		return
	default:
	}
	pw.events <- e
}

func (pw *processWatcher) Start() error {
	sock, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_DGRAM, syscall.NETLINK_CONNECTOR)
	if err != nil {
		return err
	}
	addr := &syscall.SockaddrNetlink{Family: syscall.AF_NETLINK, Pid: uint32(os.Getpid()), Groups: 1}
	if err := syscall.Bind(sock, addr); err != nil {
		return err
	}
	var op uint32 = 1
	headerSize := uint64(unsafe.Sizeof(syscall.NlMsghdr{}))
	opSize := uint64(unsafe.Sizeof(op))
	msgSize := uint64(unsafe.Sizeof(cn_msg{}))
	var iovecs = make([]unix.Iovec, 3)
	s := uint32(msgSize + opSize + headerSize)
	h := syscall.NlMsghdr{
		Len:  s,
		Pid:  uint32(os.Getegid()),
		Type: syscall.NLMSG_DONE,
	}
	iovecs[0].Base = (*byte)(unsafe.Pointer(&h))
	iovecs[0].Len = headerSize
	m := cn_msg{
		id: cb_id{
			idx: CN_IDX_PROC,
			val: CN_VAL_PROC,
		},
		len: uint16(opSize),
	}
	iovecs[1].Base = (*byte)(unsafe.Pointer(&m))
	iovecs[1].Len = msgSize
	iovecs[2].Base = (*byte)(unsafe.Pointer(&op))
	iovecs[2].Len = opSize
	_, err = unix.Writev(sock, iovecs2bytes(iovecs))
	if err != nil {
		return err
	}
	pw.sock = sock
	go func() {
		for {
			select {
			case <-pw.stop:
				return
			default:
			}
			bs := make([]byte, 1024)
			_, _, err := syscall.Recvfrom(sock, bs, 0)
			if err != nil {
				pw.sendEvents(WatchEvent{Err: err})
				return
			}
			h := (*syscall.NlMsghdr)(unsafe.Pointer(&bs[0]))
			minDataLen := uint32(headerSize) + uint32(msgSize) + uint32(unsafe.Sizeof(ProcEventHeader{}))
			if h.Len < minDataLen {
				pw.sendEvents(WatchEvent{Err: fmt.Errorf("data len %d is lower than required", h.Len)})
				continue
			}
			msg := (*cn_msg)(unsafe.Pointer(&bs[headerSize]))
			pe := ProcEvent{ptr: unsafe.Pointer(&msg.data[0])}
			pw.sendEvents(WatchEvent{ProcEvent: pe})
		}
	}()
	return nil
}
