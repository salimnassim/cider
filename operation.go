package cider

type StoreOperation string

const (
	OperationSet = StoreOperation("SET")
	OperationGet = StoreOperation("GET")
	OperationDel = StoreOperation("DEL")
)

type Operation struct {
	Name  StoreOperation
	Keys  []string
	Value string
}
