package cider

type StoreOperation string

const (
	OperationSet    = StoreOperation("SET")
	OperationGet    = StoreOperation("GET")
	OperationDel    = StoreOperation("DEL")
	OperationExists = StoreOperation("EXISTS")
	OperationExpire = StoreOperation("EXPIRE")
)

type Operation struct {
	Name  StoreOperation
	Keys  []string
	Value string
}
