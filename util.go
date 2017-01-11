package tap0901

import (
	"golang.org/x/sys/windows/registry"
	"fmt"
)

const (
	TAPWIN32_MAX_REG_SIZE = 256
	TUNTAP_COMPONENT_ID = "tap0901"
	ADAPTER_KEY = `SYSTEM\CurrentControlSet\Control\Class\{4D36E972-E325-11CE-BFC1-08002BE10318}`
	NETWORK_CONNECTIONS_KEY = `SYSTEM\CurrentControlSet\Control\Network\{4D36E972-E325-11CE-BFC1-08002BE10318}`
	USERMODEDEVICEDIR = `\\.\Global\`
	SYSDEVICEDIR = `\Device\`
	USERDEVICEDIR = `\DosDevices\Global`
	TAP_WIN_SUFFIX = ".tap"
)

const (
	TAP_WIN_IOCTL_GET_MAC = 1
	TAP_WIN_IOCTL_GET_VERSION = 2
	TAP_WIN_IOCTL_GET_MTU = 3
	TAP_WIN_IOCTL_GET_INFO = 4
	TAP_WIN_IOCTL_CONFIG_POINT_TO_POINT = 5
	TAP_WIN_IOCTL_SET_MEDIA_STATUS = 6
	TAP_WIN_IOCTL_CONFIG_DHCP_MASQ = 7
	TAP_WIN_IOCTL_GET_LOG_LINE = 8
	TAP_WIN_IOCTL_CONFIG_DHCP_SET_OPT = 9
	TAP_WIN_IOCTL_CONFIG_TUN = 10
)

const (
	FILE_ANY_ACCESS = 0
	METHOD_BUFFERED = 0
)

var (
	componentId string
)

func ctl_code(device_type, function, method, access uint32) uint32 {
	return (device_type << 16) | (access << 14) | (function << 2) | method
}

func tap_control_code(request, method uint32) uint32 {
	return ctl_code(34, request, method, FILE_ANY_ACCESS)
}

func tap_ioctl(cmd uint32) uint32 {
	return tap_control_code(cmd, METHOD_BUFFERED)
}

func matchKey(zones registry.Key, kName string, componentId string) (string, error) {
	k, err := registry.OpenKey(zones, kName, registry.READ)
	if err != nil {
		return "", err
	}
	defer k.Close()

	cId, _, err := k.GetStringValue("ComponentId")
	if cId == componentId {
		netCfgInstanceId, _, err := k.GetStringValue("NetCfgInstanceId")
		if err != nil {
			return "", err
		}
		return netCfgInstanceId, nil

	}
	return "", fmt.Errorf("ComponentId != componentId")
}

func getTuntapComponentId() (string, error) {
	if (componentId != "") {
		return componentId, nil
	}

	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		ADAPTER_KEY,
		registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	names, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return "", err
	}

	for _, name := range names {
		n, _ := matchKey(k, name, TUNTAP_COMPONENT_ID)
		if n != "" {
			componentId = n
			return n, nil
		}
	}
	return "", fmt.Errorf("Not Found")
}
