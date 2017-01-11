# GO-Tap0901库

## 说明

tap0901是openVPN项目在windows下用于替代linux下TAP/TUN设备来实现虚拟网卡
的设备。本项目致力于做一个开发者友好的TAP0901的Go库，方便制作网游加速器、虚拟
局域网、防火墙、VPN、游戏对战平台等功能。

## 依赖

请先安装OpenVPN来获得Tap0901设备，产品环境下请单独编译tap-windows6驱动避免冲突。

## 示例

```go
    // 打开TUN设备
    tun, err := tap0901.OpenTun(net.IP([]byte{0,0,0,0}), net.IP([]byte{0,0,0,0}), net.IP([]byte{0,0,0,0}))
    if err != nil {
        panic(err)
    }

    // 设置DHCP
    err = tun.SetDHCPMasq(net.IP([]byte{162, 169, 228, 206}), net.IP([]byte{255, 255, 255, 0}),
        net.IP([]byte{0, 0, 0, 0}), net.IP([]byte{0, 0, 0, 0}))
    if err != nil {
        t.Error(err)
    }

    // 得到网络名称（以太网1之类）
    szName := tun.GetNetworkName(false)

    // 启动设备
    tun.Connect()
```

其他示例参见/examples