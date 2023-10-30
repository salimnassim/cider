package cider

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

type Operation struct {
	Name  StoreOperation
	Keys  []string
	Value string
}

// Shorthand for returning an empty operation in case an error happens.
func emptyOperation() Operation {
	return Operation{
		Name:  "",
		Keys:  []string{},
		Value: "",
	}
}
