package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
)

func nextForignIp(data []byte) (remoteIp string, moreData []byte) {
	var i int
	var b byte
	for i, b = range data {
		if b == 0xA { // get to end of line
			break
		}
	}
	data = data[i:]
	for i, b = range data {
		if b == 0x3A { // get to colon after sl
			break
		}
	}
	if len(data) < i+20 {
		return
	}
	data = data[i+1:]
	for i, b = range data {
		if b == 0x3A { // get to colon after localIp
			break
		}
	}
	if len(data) < i+20 {
		return
	}
	data = data[i+6:]

	for i, b = range data {
		if b == 0x3A { // get to colon after remoteIp
			break
		}
	}
	moreData = data[i:]
	ipstr := string(data[:i])
	if len(ipstr) == 8 {

		v, err := strconv.ParseUint(ipstr, 16, 32)
		if err != nil {
			//fmt.Printf("err %v\n", err)
			return
		}
		ip := make(net.IP, net.IPv4len)
		binary.LittleEndian.PutUint32(ip, uint32(v))

		remoteIp = ip.String()
		return
	}
	if len(ipstr) == 32 {
		ip := make(net.IP, net.IPv6len)
		const grpLen = 4
		i, j := 0, 4
		for len(ipstr) != 0 {
			grp := ipstr[0:8]
			u, err := strconv.ParseUint(grp, 16, 32)
			binary.LittleEndian.PutUint32(ip[i:j], uint32(u))
			if err != nil {
				//fmt.Printf("err %v\n", err)
				return
			}
			i, j = i+grpLen, j+grpLen
			ipstr = ipstr[8:]
		}

		remoteIp = ip.String()
	}
	return
}

// 00000000:00000000:FFFF0000:010011AC
// 0000:0000:0000:0000:0000:ffff:ac11:0001
func periodicalNet(protocol string, m map[string]uint32) {
	data, err := ioutil.ReadFile(protocol)
	if err != nil {
		//fmt.Println(err)
		return
	}
	var ip string
	ip, data = nextForignIp(data)
	for data != nil {
		//fmt.Printf("Adding %s\n", ip)
		v, found := m[ip]
		if found {
			m[ip] = v + 1
		} else {
			m[ip] = 1

		}
		ip, data = nextForignIp(data)
	}
}

func periodical() {
	m := make(map[string]uint32)
	/*
		periodicalNet("/tmp/proc/net/tcp", m)
		periodicalNet("/tmp/proc/net/udp", m)
		periodicalNet("/tmp/proc/net/udplite", m)
		periodicalNet("/tmp/proc/net/tcp6", m)
		periodicalNet("/tmp/proc/net/udp6", m)
		periodicalNet("/tmp/proc/net/udplite6", m)
	*/
	periodicalNet("/proc/net/tcp", m)
	periodicalNet("/proc/net/udp", m)
	periodicalNet("/proc/net/udplite", m)
	periodicalNet("/proc/net/tcp6", m)
	periodicalNet("/proc/net/udp6", m)
	periodicalNet("/proc/net/udplite6", m)

	fmt.Printf("All IPs - %v\n", m)
}

func main() {
	periodical()
}
