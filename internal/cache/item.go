package cache

type Item[V any] struct {
	ExpireAt int64
	Value    V
}

func NewItem[V any](value V, expireAt int64) *Item[V] {
	return &Item[V]{
		ExpireAt: expireAt,
		Value:    value,
	}
}
