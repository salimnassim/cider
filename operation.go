package cider

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

type opExists struct {
	keys []string
}

type opExpire struct {
	key string
	ttl int64
	nx  bool
	xx  bool
	gt  bool
	lt  bool
}

type opIncr struct {
	key string
}

type opDecr struct {
	key string
}
