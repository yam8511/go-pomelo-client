package codec

import "errors"

// ErrPacketSizeExcced is the error used for encode/decode.
var ErrPacketSizeExcced = errors.New("codec: packet size exceed")
