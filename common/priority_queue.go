package common

type Backend struct {
	Url               string
	ActiveConnections int
}

type Backends struct {
	CurrentIndex int
	Backends     []Backend
}

func NextRoundRobin(b *Backends) Backend {
	res := b.Backends[b.CurrentIndex]
	b.CurrentIndex = (b.CurrentIndex + 1) % len(b.Backends)

	return res
}
