package tap0901

import (
	"syscall"
	"golang.org/x/sys/windows"
	"unsafe"
	"errors"
	"sync"
)

const IO_CYCLE_TIME_OUT = 100

var (
	kernel32 syscall.Handle
	waitForMutipleObjects uintptr
)

func init() {
	var err error
	kernel32, err = syscall.LoadLibrary("kernel32")
	if err != nil {
		panic(err)
	}
	waitForMutipleObjects, err = syscall.GetProcAddress(kernel32, "WaitForMultipleObjects")
	if err != nil {
		panic(err)
	}
}

func WaitForMultipleObjects(nCount uint32, handles *syscall.Handle, waitAll bool, milliseconds uint32) (uint32, error) {
	var dwWaitAll uintptr
	if waitAll {
		dwWaitAll = 1
	} else {
		dwWaitAll = 0
	}
	ret, _, err := syscall.Syscall6(waitForMutipleObjects, 4, uintptr(nCount), uintptr(unsafe.Pointer(handles)),
		dwWaitAll, uintptr(milliseconds), 0, 0)
	return uint32(ret), err
}

type event struct {
	hev syscall.Handle
	buff []byte
}

func (tun *Tun) SetReadHandler(handler func (tun *Tun, data []byte)) error {
	if tun.listening {
		return errors.New("tun already listenning")
	}
	tun.readHandler = handler
	return nil
}

func (tun *Tun) Write(data []byte) error {
	if !tun.listening {
		return errors.New("tun is not listenning")
	}
	var l uint32
	return syscall.WriteFile(tun.FD, data, &l, &tun.reusedOverlapped)
}

func (tun *Tun) postReadRequest() error {
	hevent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return err
	}
	ev := event{
		hev:     (syscall.Handle)(hevent),
		buff:    make([]byte, tun.GetMTU(false)),
	}
	tun.readReqs <- ev

	overlapped := syscall.Overlapped{}
	overlapped.HEvent = ev.hev
	var l uint32
	return syscall.ReadFile(tun.FD, ev.buff, &l, &overlapped)
}

func (tun *Tun) Worker() {
	for tun.listening {
		if err := tun.postReadRequest(); err != nil {
		}

		select {
		case data := <-tun.received:
			tun.readHandler(tun, data)
		case <- tun.closeWorker:
			break
		}
	}
}

func (tun *Tun) SignalStop() error {
	if !tun.listening {
		return errors.New("tun is not listenning")
	}
	tun.listening = false
	for i := 0; i < tun.procs; i++ {
		tun.closeWorker <- true
	}
	return nil
}

func (tun *Tun) Listen(procs int) error {
	tun.listening = true
	tun.procs = procs
	var wp sync.WaitGroup

	revents := make([]syscall.Handle, 0)
	evs := make([]event, 0)

	start := sync.WaitGroup{}
	for i := 0; i < procs; i++ {
		start.Add(1)
		go func () {
			start.Done()
			tun.Worker()
			wp.Done()
		} ()
		wp.Add(1)
	}
	defer wp.Wait()
	start.Wait()

	for tun.listening {
		SELECT:
		for {
			select{
			case ev := <-tun.readReqs:
				revents = append(revents, ev.hev)
				evs = append(evs,         ev)
			default:
				if len(revents) != 0 {
					break SELECT
				}
			}
		}

		var e uint32

		e, _ = WaitForMultipleObjects(uint32(len(revents)), &revents[0], false, IO_CYCLE_TIME_OUT)

		switch e {
		case syscall.WAIT_FAILED:
			return errors.New("wait failed")
		case syscall.WAIT_TIMEOUT:
			continue
		default:
			nIndex := e - syscall.WAIT_OBJECT_0
			tun.received <- evs[nIndex].buff

			evs = append(evs[0:nIndex], evs[nIndex+1:]...)
			revents = append(revents[0:nIndex], revents[nIndex+1 :]...)
		}
	}
	return nil
}