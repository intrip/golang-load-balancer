package common

// TODO implement an interface and multiple handlers
func (b *Backends) Next() Backend {
	res := b.Backends[b.LastSelectedIndex]
	if b.LastSelectedIndex < len(b.Backends)-1 {
		b.LastSelectedIndex++
	} else {
		b.LastSelectedIndex = 0
	}

	return res
}
