package common

type Backend struct {
	Ip                string
	Port              string
	ActiveConnections int
}

type Backends struct {
	LastSelectedIndex int
	Backends          []Backend
}

func NextRoundRobin(b *Backends) Backend {
	res := b.Backends[b.LastSelectedIndex]
	if b.LastSelectedIndex < len(b.Backends)-1 {
		b.LastSelectedIndex++
	} else {
		b.LastSelectedIndex = 0
	}

	return res
}
