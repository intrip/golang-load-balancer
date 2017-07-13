package common

// TODO test
// and then implement an interface and multiple handlers
func (b *Backends) Next() (Backend, error) {
	for _, backend := range b.Backends {
		if backend == b.LastSelected {
			return backend, nil
		}
	}
	return b.Backends[0], nil
}
