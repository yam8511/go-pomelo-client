package message

import "errors"

/**
 * ==========================
 *          訊息錯誤類型
 * ==========================
 *
 * ErrWrongMessageType 錯誤的訊息類型
 * ErrInvalidMessage 無效訊息
 * ErrRouteInfoNotFound 找不到路由
 *
 */
var (
	ErrWrongMessageType  = errors.New("wrong message type")
	ErrInvalidMessage    = errors.New("invalid message")
	ErrRouteInfoNotFound = errors.New("route info not found in dictionary")
)
