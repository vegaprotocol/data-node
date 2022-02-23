package sqlstore

type Pagination struct {
	Offset     uint64
	Limit      uint64
	Descending bool
}
