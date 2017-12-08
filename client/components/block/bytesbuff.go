package block

import (
	"bytes"

	"github.com/zero-os/0-stor/client/metastor"
)

// BytesBuffer wraps bytes.Buffer to conform to block.Writer interface
type BytesBuffer struct {
	*bytes.Buffer
}

// NewBytesBuffer creates new BytesBuffer
func NewBytesBuffer() *BytesBuffer {
	return &BytesBuffer{
		Buffer: new(bytes.Buffer),
	}
}

// WriteBlock implements block.Writer interface
func (bb *BytesBuffer) WriteBlock(key, val []byte, md *metastor.Data) (*metastor.Data, error) {
	_, err := bb.Buffer.Write(val)
	return md, err
}
