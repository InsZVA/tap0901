package tap0901

import (
	"syscall"
	"net"
	"encoding/binary"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows"
	"fmt"
)

const IO_BUFFER_NUM = 1024
const MAX_PROCS = 1024

type Tun struct {
	ID               string
	MTU              uint32
	DevicePath       string
	FD               syscall.Handle
	NetworkName      string
	received         chan []byte
	toSend           chan []byte
	readReqs         chan event
	reusedOverlapped syscall.Overlapped // reuse for write
	reusedEvent syscall.Handle
	listening        bool
	readHandler      func(tun *Tun, data []byte)
	closeWorker chan bool
	procs int
}

// OpenTun function open the tap0901 device and set config
// Params: addr -> the localIPAddr
//         network -> remoteNetwork
//         mask -> remoteNetmask
// The function configure a network for later actions
// The tun will process those transmit between local ip
// and remote network
func OpenTun(addr, network, mask net.IP) (*Tun, error) {
	id, err := getTuntapComponentId()
	if err != nil {
		return nil, err
	}

	reusedE, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return nil, err
	}
	tun := &Tun{
		ID:         id,
		DevicePath: fmt.Sprintf(USERMODEDEVICEDIR+"%s"+TAP_WIN_SUFFIX, id),
		received:   make(chan []byte, IO_BUFFER_NUM),
		toSend:     make(chan []byte, IO_BUFFER_NUM),
		readReqs:   make(chan event, IO_BUFFER_NUM),
		closeWorker: make(chan bool, MAX_PROCS),
		procs: 0,
		reusedEvent: syscall.Handle(reusedE),
	}
	tun.reusedOverlapped.HEvent = tun.reusedEvent

	fName := syscall.StringToUTF16(tun.DevicePath)
	tun.FD, err = syscall.CreateFile(
		&fName[0],
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_SYSTEM|syscall.FILE_FLAG_OVERLAPPED,
		0,
	)
	if err != nil {
		return nil, err
	}

	var returnLen uint32
	configTunParam := append(addr.To4(), network.To4()...)
	configTunParam = append(configTunParam, mask.To4()...)
	err = syscall.DeviceIoControl(tun.FD, tap_ioctl(TAP_WIN_IOCTL_CONFIG_TUN),
		&configTunParam[0], uint32(len(configTunParam)),
		&configTunParam[0], uint32(len(configTunParam)), // I think here can be nil
		&returnLen, nil)
	if err != nil {
		return nil, err
	}

	return tun, nil
}

func (tun *Tun) GetMTU(refresh bool) (uint32) {
	if !refresh && tun.MTU != 0 {
		return tun.MTU
	}

	var returnLen uint32
	var umtu = make([]byte, 4)
	err := syscall.DeviceIoControl(tun.FD, tap_ioctl(TAP_WIN_IOCTL_GET_MTU),
		&umtu[0], uint32(len(umtu)),
		&umtu[0], uint32(len(umtu)),
		&returnLen, nil)
	if err != nil {
		return 0
	}
	tun.MTU = binary.LittleEndian.Uint32(umtu)

	return tun.MTU
}

func (tun *Tun) Connect() error {
	var returnLen uint32
	inBuffer := []byte("\x01\x00\x00\x00") // only means TRUE
	err := syscall.DeviceIoControl(
		tun.FD, tap_ioctl(TAP_WIN_IOCTL_SET_MEDIA_STATUS),
		&inBuffer[0], uint32(len(inBuffer)),
		&inBuffer[0], uint32(len(inBuffer)),
		&returnLen, nil)
	return err
}

func (tun *Tun) SetDHCPMasq(dhcpAddr, dhcpMask, serverIP, leaseTime net.IP) error {
	var returnLen uint32
	configTunParam := append(dhcpAddr.To4(), dhcpMask.To4()...)
	configTunParam = append(configTunParam, serverIP.To4()...)
	configTunParam = append(configTunParam, leaseTime.To4()...)
	err := syscall.DeviceIoControl(tun.FD, tap_ioctl(TAP_WIN_IOCTL_CONFIG_DHCP_MASQ),
		&configTunParam[0], uint32(len(configTunParam)),
		&configTunParam[0], uint32(len(configTunParam)), // I think here can be nil
		&returnLen, nil)
	return err
}

func (tun *Tun) GetNetworkName(refresh bool) string {
	if !refresh && tun.NetworkName != "" {
		return tun.NetworkName
	}
	keyName := `SYSTEM\CurrentControlSet\Control\Network\{4D36E972-E325-11CE-BFC1-08002BE10318}\` +
		tun.ID + `\Connection`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyName, registry.ALL_ACCESS)
	if err != nil {
		return ""
	}
	szname, _, err := k.GetStringValue("Name")
	if err != nil {
		return ""
	}
	k.Close()
	tun.NetworkName = szname
	return szname
}
