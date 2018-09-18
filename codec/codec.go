package codec

import (
	"bytes"

	"github.com/yam8511/go-pomelo-client/packet"
)

// NewDecoder returns a new decoder that used for decode network bytes slice.
func NewDecoder() *Decoder {
	return &Decoder{
		buf:  bytes.NewBuffer(nil),
		size: -1,
	}
}

// A Decoder reads and decodes network data slice
type Decoder struct {
	buf  *bytes.Buffer
	size int  // last packet length
	typ  byte // last packet type
}

func (c *Decoder) forward() error {
	header := c.buf.Next(HeadLength)
	c.typ = header[0]
	if c.typ < packet.Handshake || c.typ > packet.Kick {
		return packet.ErrWrongPacketType
	}
	c.size = bytesToInt(header[1:])

	// packet length limitation
	if c.size > MaxPacketSize {
		return ErrPacketSizeExcced
	}
	return nil
}

// Decode decode the network bytes slice to packet.Packet(s)
// TODO(Warning): shared slice
func (c *Decoder) Decode(data []byte) ([]*packet.Packet, error) {
	c.buf.Write(data)

	var (
		packets []*packet.Packet
		err     error
	)
	// check length
	if c.buf.Len() < HeadLength {
		return nil, err
	}

	// first time
	if c.size < 0 {
		if err = c.forward(); err != nil {
			return nil, err
		}
	}

	for c.size <= c.buf.Len() {
		p := &packet.Packet{Type: byte(c.typ), Length: c.size, Data: c.buf.Next(c.size)}
		packets = append(packets, p)

		// more packet
		if c.buf.Len() < HeadLength {
			c.size = -1
			break
		}

		if err = c.forward(); err != nil {
			return nil, err

		}

	}

	return packets, nil
}

// Encode create a packet.Packet from  the raw bytes slice and then encode to network bytes slice
// Protocol refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
//
// -<type>-|--------<length>--------|-<data>-
// --------|------------------------|--------
// 1 byte packet type, 3 bytes packet data length(big end), and data segment
func Encode(typ byte, data []byte) ([]byte, error) {
	if typ < packet.Handshake || typ > packet.Kick {
		return nil, packet.ErrWrongPacketType
	}

	dataLen := len(data)
	buf := make([]byte, dataLen+HeadLength)
	buf[0] = byte(typ)

	// log.Println("Package 類型 ---> ", typ)
	// log.Println("Package 類型的 buffer ---> ", buf)
	copy(buf[1:HeadLength], intToBytes(dataLen))
	// log.Println("Header長度 ---> ", dataLen)
	// log.Println("Header長度的 buffer ---> ", buf[1:HeadLength])
	copy(buf[HeadLength:], data)
	// log.Println("本身資料的 buffer ---> ", buf[HeadLength:])

	return buf, nil
}
