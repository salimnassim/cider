package cider

import (
	"strings"
)

// Parses a client command and returns the operation and all arguments.
func ParseCommand(cmd []byte) Operation {
	// split fields to get the operation
	fields := strings.Fields(string(cmd))
	if len(fields) == 0 {
		return Operation{
			Name:  "",
			Keys:  []string{},
			Value: "",
		}
	}

	// parsed operation
	switch po := strings.ToUpper(fields[0]); po {
	case "SET":
		return Operation{
			Name:  StoreOperation(po),
			Keys:  []string{fields[1]},
			Value: strings.Join(fields[2:], ""),
		}
	case "GET":
		return Operation{
			Name:  StoreOperation(po),
			Keys:  []string{fields[1]},
			Value: "",
		}
	case "DEL":
		return Operation{
			Name:  StoreOperation(po),
			Keys:  fields[1:],
			Value: "",
		}
	case "EXISTS":
		return Operation{
			Name:  StoreOperation(po),
			Keys:  fields[1:],
			Value: "",
		}
	case "EXPIRE":
		return Operation{
			Name:  StoreOperation(po),
			Keys:  []string{fields[1]},
			Value: fields[2],
		}
	default:
		return Operation{
			Name:  "",
			Keys:  []string{},
			Value: "",
		}
	}
}
