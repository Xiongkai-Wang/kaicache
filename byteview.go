package kaicache

// A ByteView holds an immutable(unchangable) view of bytes
type ByteView struct {
	b []byte // 选择 byte 类型是为了能够支持任意的数据类型的存储，例如字符串、图片等
}

func (v ByteView) Len() int { // implement Value
	return len(v.b)
}

// b can only be read; prevent overwrite from outside program
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
