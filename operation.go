package cider

type StoreKey string
type StoreOperation string

const (
	OperationSet    = StoreOperation("SET")
	OperationGet    = StoreOperation("GET")
	OperationDel    = StoreOperation("DEL")
	OperationExists = StoreOperation("EXISTS")
	OperationExpire = StoreOperation("EXPIRE")
	OperationIncr   = StoreOperation("INCR")
	OperationDecr   = StoreOperation("DECR")
)

type opSet struct {
	key   string
	value []byte
	nx    bool
	xx    bool
	get   bool
	ex    int64
	// px      int64
	exat int64
	// pxat    int64
	keepttl bool
}

type opGet struct {
	key string
}

type opDel struct {
	keys []string
}

type opExpire struct {
	key string
	ttl int64
}

type operation struct {
	name  StoreOperation
	keys  []string
	value string
}

// Shorthand for returning an empty operation in case an error happens.
func emptyOperation() operation {
	return operation{
		name:  "",
		keys:  []string{},
		value: "",
	}
}
