package store

// Store interfce defines the generic interface for possible datastores to implement.
type Store interface {
	Put(key string, value any) error
	Get(key string) (any, error)
	Count() (int, error)
	List() (any, error)
}
