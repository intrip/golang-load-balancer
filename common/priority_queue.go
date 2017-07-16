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
	b.LastSelectedIndex = (b.LastSelectedIndex + 1) % len(b.Backends)

	return res
}
