package natter

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"net"
	"time"
)

const (
	magicCookie            = 0x2112A442
	fingerprint            = 0x5354554e
	typeBindingRequest     = 0x0001
	attributeSoftware      = 0x8022
	attributeFingerprint   = 0x8028
	attributeChangeRequest = 0x0003
)

type attribute struct {
	types  uint16
	length uint16
	value  []byte
}

// padding 将字节数组填充到大于或等于字节数组长度的最小 4 的倍数
func padding(bytes []byte) []byte {
	length := uint16(len(bytes))
	return append(bytes, make([]byte, align(length)-length)...)
}

// newAttribute 创建一个新的属性
func newAttribute(types uint16, value []byte) attribute {
	att := new(attribute)
	att.types = types
	att.value = padding(value)
	att.length = uint16(len(att.value))
	return *att
}

// newChangeReqAttribute 创建一个新的改变请求属性
func newChangeReqAttribute(changeIP bool, changePort bool) attribute {
	value := make([]byte, 4)
	if changeIP {
		value[3] |= 0x04
	}
	if changePort {
		value[3] |= 0x02
	}
	return newAttribute(attributeChangeRequest, value)
}

type packet struct {
	types      uint16
	length     uint16
	transID    []byte // 4 bytes magic cookie + 12 bytes transaction id
	attributes []attribute
}

// newPacket 创建一个新的数据包
func newPacket() (*packet, error) {
	v := new(packet)
	v.transID = make([]byte, 16)                           // 4 bytes magic cookie + 12 bytes transaction id
	binary.BigEndian.PutUint32(v.transID[:4], magicCookie) // 魔法Cookie
	_, err := rand.Read(v.transID[4:])                     // 事务ID
	if err != nil {
		return nil, err
	}
	v.attributes = make([]attribute, 0, 10) // 10 个属性
	v.length = 0                            // 长度
	return v, nil
}

// align 将 uint16 数字对齐到大于或等于 uint16 数字的最小 4 的倍数
func align(n uint16) uint16 {
	return (n + 3) & 0xfffc
}
func (v *packet) addAttribute(a attribute) {
	v.attributes = append(v.attributes, a)
	v.length += align(a.length) + 4
}

func (v *packet) bytes() []byte {
	packetBytes := make([]byte, 4)
	binary.BigEndian.PutUint16(packetBytes[0:2], v.types)
	binary.BigEndian.PutUint16(packetBytes[2:4], v.length)
	packetBytes = append(packetBytes, v.transID...)
	for _, a := range v.attributes {
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, a.types)
		packetBytes = append(packetBytes, buf...)
		binary.BigEndian.PutUint16(buf, a.length)
		packetBytes = append(packetBytes, buf...)
		packetBytes = append(packetBytes, a.value...)
	}
	return packetBytes
}

// GetMapping 获取 STUN 映射
func GetMapping(
	specifyLocalAddress *net.TCPAddr,
	// changeIP, changePort bool,
) (lAddr, rAddr *net.TCPAddr, err error) {
	var conn *net.TCPConn
	var l, r *net.TCPAddr
	for _, stunServer := range stunList {
		// 判断 conn 是否已经关闭
		if err != nil {
			logger.Error(fmt.Sprintf("获取 STUN 映射失败: %s\n", err.Error()))
		}
		// 判断是否有端口
		if _, _, err = net.SplitHostPort(stunServer); err != nil {
			stunServer = fmt.Sprintf("%s:3478", stunServer)
			err = nil
		}
		if specifyLocalAddress != nil {
			l = specifyLocalAddress
		} else {
			l, err = net.ResolveTCPAddr("tcp4", "0.0.0.0:0")
		}
		if err != nil {
			continue
		}
		r, err = net.ResolveTCPAddr("tcp4", stunServer)
		if err != nil {
			continue
		}
		conn, err = net.DialTCP("tcp4", l, r)
		if err != nil {
			continue
		}
		//message := make([]byte, 20)
		//binary.BigEndian.PutUint32(message[0:], 0x00010000)         // 消息类型和长度
		//binary.BigEndian.PutUint32(message[4:], magicCookie)        // 魔法Cookie
		//binary.BigEndian.PutUint32(message[8:], 0x4E415452)         // 事务ID
		//binary.BigEndian.PutUint32(message[12:], mathRand.Uint32()) // 事务ID
		//binary.BigEndian.PutUint32(message[16:], mathRand.Uint32()) // 事务ID
		var pkt *packet
		pkt, err = newPacket() // 创建新的数据包 设置事务ID 魔法Cookie
		if err != nil {
			continue
		}
		pkt.types = typeBindingRequest                                     // 绑定请求
		attribute := newAttribute(attributeSoftware, []byte("StunClient")) // 软件属性

		//if changeIP || changePort {
		//	pkt.addAttribute(newChangeReqAttribute(changeIP, changePort)) // 改变请求属性
		//}

		pkt.addAttribute(attribute)                          // 添加属性
		pkt.length += 8                                      // 长度增加 8
		crc := crc32.ChecksumIEEE(pkt.bytes()) ^ fingerprint // 计算 CRC
		buf := make([]byte, 4)                               // 创建 4 字节的缓冲区
		binary.BigEndian.PutUint32(buf, crc)                 // 将 CRC 写入缓冲区
		attribute = newAttribute(attributeFingerprint, buf)  // 指纹属性
		pkt.length -= 8                                      // 长度减去 8
		pkt.addAttribute(attribute)                          // 添加属性

		_, err = conn.Write(pkt.bytes())
		if err != nil {
			continue
		}
		// 设置超时时间 3s
		err = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		if err != nil {
			continue
		}
		// 读取数据
		buf = make([]byte, 1024)
		_, err = conn.Read(buf)
		if err != nil {
			continue
		}
		err = conn.Close()
		if err != nil {
			continue
		}
		payload := buf[20:]
		var ip uint32
		var port uint16
		for len(payload) > 0 {
			attrType := binary.BigEndian.Uint16(payload[:2])
			attrLen := binary.BigEndian.Uint16(payload[2:4])
			if attrType == 1 || attrType == 32 {
				port = binary.BigEndian.Uint16(payload[6:8])
				ip = binary.BigEndian.Uint32(payload[8:12])
				if attrType == 32 {
					port ^= 0x2112
					ip ^= 0x2112A442
				}
				break
			}
			payload = payload[4+attrLen:]
		}
		outerAddr := net.IPv4(byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip)).String()
		lAddr, err = net.ResolveTCPAddr("tcp4", conn.LocalAddr().String())
		if err != nil {
			continue
		}
		rAddr, err = net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", outerAddr, port))
		if err != nil {
			continue
		}
		return
	}
	return
}
