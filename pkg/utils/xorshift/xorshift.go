package Xorshift

// Xorshift 生成器类型
type Xorshift struct {
	state uint32
}

// NewXorshift 创建一个新的 Xorshift 实例
func NewXorshift(seed uint32) *Xorshift {
	return &Xorshift{state: seed}
}

// Intn 生成 0 到 n-1 之间的随机数
func (x *Xorshift) Intn(n int) int {
	if n <= 0 {
		panic("n must be positive")
	}
	return int(x.xorshift() % uint32(n))
}

// xorshift 生成一个随机数
func (x *Xorshift) xorshift() uint32 {
	x.state ^= x.state << 13
	x.state ^= x.state >> 17
	x.state ^= x.state << 5
	return x.state
}
