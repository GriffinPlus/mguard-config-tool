package atv

// ValueType is the value type of a setting in an ATV file.
type ValueType int

const (
	// CIDR - Network/IP-Address in CIDR notation (e.g. 192.168.1.0/24)
	CIDR ValueType = iota
	// HEX - A hexadecimal value
	HEX
	// IP - An IP address (e.g. 192.168.1.2)
	IP
	// MAC - A MAC address (e.g. 00:0c:be:12:fe:01)
	MAC
	// NETMASK - A subnet mask (e.g. 255.255.255.0)
	NETMASK
	// NUM - A numeric value
	NUM
	// TXT - A textual value
	TXT
	// ROWREF - A reference id of a defined row (e.g. MAI0983174920)
	ROWREF
)

func (d ValueType) String() string {
	return [...]string{"CIDR", "HEX", "IP", "MAC", "NETMASK", "NUM", "TXT", "ROWREF"}[d]
}
