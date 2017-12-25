package consensus

type PolicyLevel byte

const (
	AllowAll PolicyLevel = 0x00
	DenyAll PolicyLevel = 0x01
	AllowList PolicyLevel = 0x02
	DenyList PolicyLevel = 0x03
)
