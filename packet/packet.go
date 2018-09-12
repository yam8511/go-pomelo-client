package packet

import (
	"fmt"
)

// [Reference](https://github.com/NetEase/pomelo/wiki/Communication-Protocol)

// New 建立一個空的Packet
func New() *Packet {
	return &Packet{}
}

// Packet 定義網路包格式
type Packet struct {
	Type   byte // 詳見 constan.go
	Length int  // body内容长度，3个byte的大端整数，因此最大的包长度为2^24个byte。
	Data   []byte
}

// String 以文字顯示Packet資料
func (p *Packet) String() string {
	return fmt.Sprintf("Type: %d, Length: %d, Data: %s", p.Type, p.Length, string(p.Data))
}
