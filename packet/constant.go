package packet

// [Reference](https://github.com/NetEase/pomelo/wiki/Communication-Protocol)

/**
 * ==========================
 *        Packet 類型
 * ==========================
 *
 * Handshake 客户端到服務器的握手请求以及服務器到客户端的握手回應
 * HandshakeAck 客户端到服務器的握手ack
 * Heartbeat 心跳包
 * Data 數據包
 * Kick 服務器主動斷開連接通知
 *
 */
const (
	Handshake    = 0x01
	HandshakeAck = 0x02
	Heartbeat    = 0x03
	Data         = 0x04
	Kick         = 0x05
)
