package common

type PriorityQueuer interface {
	Next() Backend
}

type Backend struct {
	Url               string
	ActiveConnections int
}

type RoundRobin struct {
	CurrentIndex int
	Backends     []Backend
}

func Next(b *RoundRobin) Backend {
	res := b.Backends[b.CurrentIndex]
	b.CurrentIndex = (b.CurrentIndex + 1) % len(b.Backends)

	return res
}
