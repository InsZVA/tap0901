package main

import (
	"net"
	"sync"
	"github.com/inszva/tap0901"
	"time"
	"fmt"
	"encoding/hex"
)

func main() {
	tun, err := tap0901.OpenTun([]byte{123, 123, 123, 123}, []byte{0, 0, 0, 0}, []byte{0, 0, 0, 0})
	if err != nil {
		panic(err)
	}
	tun.Connect()
	time.Sleep(2 * time.Second)

	tun.SetReadHandler(func (tun *tap0901.Tun, data []byte) {
		fmt.Println(hex.EncodeToString(data))
	})
	wp := sync.WaitGroup{}
	wp.Add(1)
	go func () {
		tun.Listen(1)
		wp.Done()
	} ()
	time.Sleep(2 * time.Second)
	laddr, _ := net.ResolveUDPAddr("udp4", "123.123.123.123:15645")
	raddr, _ := net.ResolveUDPAddr("udp4", "123.2.3.4:25457")
	conn, err := net.DialUDP("udp4", laddr, raddr)
	if err != nil {
		panic(err)
	}
	//fmt.Println(tun.GetMTU(false))
	//fmt.Println(tun.Write([]byte{0x45, 0x00, 0x00, 0x23, 0x12, 0xa5, 0x00, 0x40, 0x11, 0xf3, 0x28, 0x7b, 0x02, 0x03, 0x04, 0x7b, 0x7b, 0x7b, 0x7b, 0x3d, 0x1d, 0x63, 0x71, 0x00, 0x0f, 0x59, 0x17, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67}))
	conn.Write([]byte("abcdefg"))
	conn.Write([]byte("abcdefg"))
	conn.Write([]byte("abcdefg"))
	//tun.Write([]byte("abcdefg"))
	time.Sleep(10*time.Second)
	tun.SignalStop()
	wp.Wait()

}
