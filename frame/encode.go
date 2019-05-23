package frame

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/negasus/haproxy-spoe-go/varint"
	"io"
)

func (f *Frame) Encode(dest io.Writer) (n int, err error) {

	buf := bytes.Buffer{}

	buf.WriteByte(byte(f.Type))

	binary.BigEndian.PutUint32(f.tmp, f.Flags)

	buf.Write(f.tmp)

	n = varint.PutUvarint(f.varintBuf, f.StreamID)
	buf.Write(f.varintBuf[:n])

	n = varint.PutUvarint(f.varintBuf, f.FrameID)
	buf.Write(f.varintBuf[:n])

	var payload []byte

	switch f.Type {
	case TypeAgentHello, TypeAgentDisconnect:
		payload, err = f.KV.Bytes()
		if err != nil {
			return
		}

	case TypeAgentAck:
		for _, act := range *f.Actions {
			payload, err = (*act).Marshal(payload)
			if err != nil {
				return
			}
		}

	default:
		err = fmt.Errorf("unexpected frame type %d", f.Type)
		return
	}

	buf.Write(payload)

	frameSizeBuf := make([]byte, 4)

	binary.BigEndian.PutUint32(frameSizeBuf, uint32(buf.Len()))

	n, err = dest.Write(frameSizeBuf)
	if err != nil || n != len(frameSizeBuf) {
		return 0, fmt.Errorf("error write frameSize. writes %d, expect %d, err: %v", n, len(frameSizeBuf), err)
	}

	n, err = dest.Write(buf.Bytes())
	if err != nil || n != buf.Len() {
		return 0, fmt.Errorf("error write frame. writes %d, expect %d, err: %v", n, len(frameSizeBuf), err)
	}

	return len(frameSizeBuf) + buf.Len(), nil
}