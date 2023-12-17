package context

type ContextKey struct {
	Name string
}

func (k *ContextKey) String() string {
	return k.Name
}
