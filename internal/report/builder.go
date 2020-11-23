package report

// Builder is a utility for constructing test reprots.
type Builder struct {
	Log func(string)
}

func (b *Builder) AddFinding(f Finding) {
}

func (b *Builder) Finalize() {
}
