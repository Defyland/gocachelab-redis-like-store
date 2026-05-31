package server

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Defyland/gocachelab-redis-like-store/internal/protocol"
	"github.com/Defyland/gocachelab-redis-like-store/internal/store"
)

func ApplyRecoveredCommand(store *store.Store, command protocol.Command) error {
	switch command.Name {
	case "SET":
		if len(command.Args) != 2 {
			return fmt.Errorf("SET replay expects 2 args")
		}
		store.Set(command.Args[0], command.Args[1], time.Time{})
		return nil
	case "DEL":
		if len(command.Args) < 1 {
			return fmt.Errorf("DEL replay expects at least 1 arg")
		}
		store.Del(command.Args...)
		return nil
	case "EXPIREAT":
		if len(command.Args) != 2 {
			return fmt.Errorf("EXPIREAT replay expects 2 args")
		}
		unixNano, err := strconv.ParseInt(command.Args[1], 10, 64)
		if err != nil {
			return err
		}
		store.ExpireAt(command.Args[0], time.Unix(0, unixNano).UTC())
		return nil
	case "PERSIST":
		if len(command.Args) != 1 {
			return fmt.Errorf("PERSIST replay expects 1 arg")
		}
		store.Persist(command.Args[0])
		return nil
	default:
		return fmt.Errorf("unsupported replay command %s", command.Name)
	}
}
