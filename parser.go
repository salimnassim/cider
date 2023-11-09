package cider

import (
	"errors"
	"strconv"
	"strings"
)

func ParseCommandAny(command []byte) (any, error) {
	fields := strings.Fields(string(command))
	if len(fields) <= 0 {
		return nil, errors.New("no command supplied")
	}

	operation := strings.ToUpper(fields[0])
	switch operation {
	// https://redis.io/commands/set/
	case "SET":
		// need at least key and value
		if len(fields) < 3 {
			return nil, errors.New("not enough arguments for SET")
		}

		// cutoff point to read the value only excluding ex/exat number
		cutoff := 256
		var op opSet
		var sb strings.Builder

		for i, v := range fields {
			if i == 1 {
				op.key = fields[1]
				continue
			}
			if i > 2 && v == "NX" {
				if op.xx {
					return nil, errors.New("XX already set in this command")
				}
				op.nx = true
				continue
			}
			if i > 2 && v == "XX" {
				if op.nx {
					return nil, errors.New("NX already set in this command")
				}
				op.xx = true
				continue
			}
			if i > 2 && v == "GET" {
				op.get = true
				continue
			}
			if i > 2 && v == "EX" {
				if op.exat != 0 {
					return nil, errors.New("EXAT already set in this command")
				}

				if len(fields) <= i+1 {
					return nil, errors.New("EX value missing")
				}

				secs, err := strconv.ParseInt(fields[i+1], 10, 64)
				if err != nil {
					return nil, errors.New("unable to parse EX number")
				}
				op.ex = secs
				cutoff = i
				continue
			}
			if i > 2 && v == "EXAT" {
				if op.ex != 0 {
					return nil, errors.New("EX already set in this command")
				}

				if len(fields) <= i+1 {
					return nil, errors.New("EXAT value missing")
				}

				ts, err := strconv.ParseInt(fields[i+1], 10, 64)
				if err != nil {
					return nil, errors.New("unable to parse EXAT timestamp")
				}
				op.exat = ts
				cutoff = i
				continue
			}
			if i > 2 && v == "KEEPTTL" {
				op.keepttl = true
				continue
			}

			if i > 1 && i < cutoff {
				sb.WriteString(v + " ")
			}
		}

		op.value = []byte(strings.TrimSpace(sb.String()))

		return op, nil

	// https://redis.io/commands/get/
	case "GET":
		if len(fields) < 2 {
			return nil, errors.New("not enough arguments for GET")
		}

		op := opGet{
			key: fields[1],
		}

		return op, nil

	// https://redis.io/commands/del/
	case "DEL":
		if len(fields) < 2 {
			return nil, errors.New("not enough arguments for DEL")
		}

		op := opDel{
			keys: fields[1:],
		}

		return op, nil

	// https://redis.io/commands/exists/
	case "EXISTS":
		if len(fields) < 2 {
			return nil, errors.New("not enough arguments for EXISTS")
		}

		op := opExists{
			keys: fields[1:],
		}

		return op, nil

	// https://redis.io/commands/expire/
	case "EXPIRE":
		// need at least key and value
		if len(fields) < 3 {
			return nil, errors.New("not enough arguments for EXPIRE")
		}

		var op opExpire

		for i, v := range fields {
			if i == 1 {
				op.key = fields[1]

				secs, err := strconv.ParseInt(fields[i+1], 10, 64)
				if err != nil {
					return nil, errors.New("unable to parse EXPIRE number")
				}
				op.ttl = secs
				continue
			}
			if i > 2 && v == "NX" {
				if op.xx || op.gt || op.lt {
					return nil, errors.New("XX/GT/LT already set in this command")
				}
				op.nx = true
				continue
			}
			if i > 2 && v == "XX" {
				if op.nx {
					return nil, errors.New("NX already set in this command")
				}
				op.xx = true
				continue
			}
			if i > 2 && v == "GT" {
				if op.nx {
					return nil, errors.New("NX already set in this command")
				}
				op.gt = true
				continue
			}
			if i > 2 && v == "LT" {
				if op.nx {
					return nil, errors.New("NX already set in this command")
				}
				op.lt = true
				continue
			}
		}

		return op, nil

	// https://redis.io/commands/incr/
	case "INCR":
		if len(fields) < 2 {
			return nil, errors.New("not enough arguments for INCR")
		}

		op := opIncr{
			key: fields[1],
		}

		return op, nil

	// https://redis.io/commands/decr/
	case "DECR":
		if len(fields) < 2 {
			return nil, errors.New("not enough arguments for DECR")
		}

		op := opDecr{
			key: fields[1],
		}

		return op, nil
	}

	return nil, errors.New("unsupported operation")
}

// Parses a client command and returns the operation and all arguments.
func ParseCommand(cmd []byte) (operation, error) {
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
		return operation{
			name:  StoreOperation(po),
			keys:  []string{fields[1]},
			value: strings.Join(fields[2:], " "),
		}, nil
	case "GET":
		if len(fields) < 2 {
			return emptyOperation(), errors.New("need more parameters (2)")
		}
		return operation{
			name:  StoreOperation(po),
			keys:  []string{fields[1]},
			value: "",
		}, nil
	case "DEL":
		if len(fields) < 2 {
			return emptyOperation(), errors.New("need more parameters (2)")
		}
		return operation{
			name:  StoreOperation(po),
			keys:  fields[1:],
			value: "",
		}, nil
	case "EXISTS":
		if len(fields) < 2 {
			return emptyOperation(), errors.New("need more parameters (2)")
		}
		return operation{
			name:  StoreOperation(po),
			keys:  fields[1:],
			value: "",
		}, nil
	case "EXPIRE":
		if len(fields) < 3 {
			return emptyOperation(), errors.New("need more parameters (3)")
		}
		return operation{
			name:  StoreOperation(po),
			keys:  []string{fields[1]},
			value: fields[2],
		}, nil
	case "INCR":
		if len(fields) < 2 {
			return emptyOperation(), errors.New("need more parameters (2)")
		}
		return operation{
			name:  StoreOperation(po),
			keys:  []string{fields[1]},
			value: "",
		}, nil
	case "DECR":
		if len(fields) < 2 {
			return emptyOperation(), errors.New("need more parameters (2)")
		}
		return operation{
			name:  StoreOperation(po),
			keys:  []string{fields[1]},
			value: "",
		}, nil
	default:
		return operation{
			name:  "",
			keys:  []string{},
			value: "",
		}, nil
	}
}
