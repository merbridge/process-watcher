package processwatcher

import (
	"unsafe"
)

type ProcEventType uint32
type Pid uint32

const (
	PROC_EVENT_NONE     ProcEventType = 0x00000000
	PROC_EVENT_FORK                   = 0x00000001
	PROC_EVENT_EXEC                   = 0x00000002
	PROC_EVENT_UID                    = 0x00000004
	PROC_EVENT_GID                    = 0x00000040
	PROC_EVENT_SID                    = 0x00000080
	PROC_EVENT_PTRACE                 = 0x00000100
	PROC_EVENT_COMM                   = 0x00000200
	PROC_EVENT_COREDUMP               = 0x40000000
	PROC_EVENT_EXIT                   = 0x80000000
)

type WatchEvent struct {
	Err error
	ProcEvent
}

type ProcEventHeader struct {
	Type ProcEventType
	CPU  uint32
	// Number of nano seconds since system boot
	Timestamps uint64
}

type ProcEvent struct {
	ptr unsafe.Pointer
}

type Fork struct {
	ParentPid  Pid
	ParentTgid Pid
	ChildPid   Pid
	ChildTgid  Pid
}

type Exec struct {
	ProcessPid  Pid
	ProcessTgid Pid
}

type Comm struct {
	ProcessPid  Pid
	ProcessTgid Pid
}

type Exit struct {
	ProcessPid  Pid
	ProcessTgid Pid
	ExitCode    uint32
	ExitSignal  uint32
	ParentPid   Pid
	ParentTgid  Pid
}

func (p *ProcEvent) GetHeader() ProcEventHeader {
	return *(*ProcEventHeader)(p.ptr)
}

func (p *ProcEvent) GetType() ProcEventType {
	return p.GetHeader().Type
}

func (p *ProcEvent) dataPtr() unsafe.Pointer {
	return unsafe.Pointer(uintptr(p.ptr) + uintptr(unsafe.Sizeof(ProcEventHeader{})))
}

func (p *ProcEvent) GetFork() Fork {
	return *(*Fork)(p.dataPtr())
}

func (p *ProcEvent) GetExec() Exec {
	return *(*Exec)(p.dataPtr())
}

func (p *ProcEvent) GetComm() Comm {
	return *(*Comm)(p.dataPtr())
}

func (p *ProcEvent) GetExit() Exit {
	return *(*Exit)(p.dataPtr())
}
