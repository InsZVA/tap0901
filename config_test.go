package tap0901

import (
	"testing"
	"net"
	"time"
)

var tun *Tun

func TestOpenTun(t *testing.T) {
	var err error
	tun, err = OpenTun(net.IP([]byte{0,0,0,0}), net.IP([]byte{0,0,0,0}), net.IP([]byte{0,0,0,0}))
	if err != nil {
		t.Error(err)
	}
}

func TestTun_SetDHCPMasq(t *testing.T) {
	err := tun.SetDHCPMasq(net.IP([]byte{162, 169, 228, 206}), net.IP([]byte{255, 255, 255, 0}),
		net.IP([]byte{0, 0, 0, 0}), net.IP([]byte{0, 0, 0, 0}))
	if err != nil {
		t.Error(err)
	}
}

func TestTun_Connect(t *testing.T) {
	err := tun.Connect()
	if err != nil {
		t.Error(err)
	}
	time.Sleep(20 * time.Second)
}