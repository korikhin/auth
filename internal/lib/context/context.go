package context

type ContextKey struct {
	Name string
}

func (k *ContextKey) String() string {
	return k.Name
}

var (
	StatusKey  = &ContextKey{"Status"}
	RequestKey = &ContextKey{"RequestID"}
	UserKey    = &ContextKey{"User"}
)

// TODO: Contexts for User, etc.
var ()
