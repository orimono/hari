package dispatcher

import (
	"log/slog"

	"github.com/orimono/hari/internal/protocol"
)

func Dispatch(msg *protocol.Message) {
	slog.Info(string(msg.Data))
}
