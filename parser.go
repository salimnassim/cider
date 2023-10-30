package cider

import (
	"errors"
	"strings"
)

// Parses a client command and returns the operation and all arguments.
func ParseCommand(cmd []byte) (Operation, error) {
	// split fields to get the operation
	fields := strings.Fields(string(cmd))
	if len(fields) == 0 {
		return emptyOperation(), errors.New("no command supplied")
	}

	// parsed operation
	switch po := strings.ToUpper(fields[0]); po {
	case "SET":
		if len(fields) < 3 {
			return emptyOperation(), errors.New("need more parameters (3)")
		}
		return Operation{
			Name:  StoreOperation(po),
			Keys:  []string{fields[1]},
			Value: strings.Join(fields[2:], " "),
		}, nil
	case "GET":
		if len(fields) < 2 {
			return emptyOperation(), errors.New("need more parameters (2)")
		}
		return Operation{
			Name:  StoreOperation(po),
			Keys:  []string{fields[1]},
			Value: "",
		}, nil
	case "DEL":
		if len(fields) < 2 {
			return emptyOperation(), errors.New("need more parameters (2)")
		}
		return Operation{
			Name:  StoreOperation(po),
			Keys:  fields[1:],
			Value: "",
		}, nil
	case "EXISTS":
		if len(fields) < 2 {
			return emptyOperation(), errors.New("need more parameters (2)")
		}
		return Operation{
			Name:  StoreOperation(po),
			Keys:  fields[1:],
			Value: "",
		}, nil
	case "EXPIRE":
		if len(fields) < 3 {
			return emptyOperation(), errors.New("need more parameters (3)")
		}
		return Operation{
			Name:  StoreOperation(po),
			Keys:  []string{fields[1]},
			Value: fields[2],
		}, nil
	case "INCR":
		if len(fields) < 2 {
			return emptyOperation(), errors.New("need more parameters (2)")
		}
		return Operation{
			Name:  StoreOperation(po),
			Keys:  []string{fields[1]},
			Value: "",
		}, nil
	case "DECR":
		if len(fields) < 2 {
			return emptyOperation(), errors.New("need more parameters (2)")
		}
		return Operation{
			Name:  StoreOperation(po),
			Keys:  []string{fields[1]},
			Value: "",
		}, nil
	default:
		return Operation{
			Name:  "",
			Keys:  []string{},
			Value: "",
		}, nil
	}
}
