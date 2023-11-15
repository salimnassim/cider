# cider
Experimental Redis clone written in Go that supports a limited subset of commands.

---

### Supported commands

Currently supports the following commands

SET, GET, DEL, EXISTS, EXPIRE, INCR, DECR, TTL

### Store limitations

Store keys have to be UTF-8 strings and values are represented as an arbitrary byte array.

#### Resources

https://redis.io/docs/reference/protocol-spec/