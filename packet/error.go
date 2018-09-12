package packet

import "errors"

/**
 * ==========================
 *        Packet錯誤類型
 * ==========================
 *
 * ErrWrongPacketType 錯誤的Packet型態
 *
 */
var (
	ErrWrongPacketType = errors.New("wrong packet type")
)
