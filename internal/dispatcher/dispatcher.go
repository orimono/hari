package dispatcher

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gorilla/websocket"
	"github.com/orimono/ito"
	"github.com/orimono/shutter/internal/capability"
	"github.com/orimono/shutter/internal/protocol"
	"github.com/orimono/shutter/internal/store"
)

type Dispatcher struct {
	executors map[string]capability.Executor
	store     *store.ExecutorStore
	manager   *capability.Manager
}

func New(mgr *capability.Manager, s *store.ExecutorStore) *Dispatcher {
	executors := make(map[string]capability.Executor, len(mgr.Executors()))
	for _, e := range mgr.Executors() {
		executors[e.Name()] = e
	}
	return &Dispatcher{
		executors: executors,
		store:     s,
		manager:   mgr,
	}
}

// Handle decodes an incoming envelope and routes it to the appropriate handler.
func (d *Dispatcher) Handle(ctx context.Context, data []byte, reply func(protocol.Message)) {
	env, err := ito.Decode(data)
	if err != nil {
		slog.Debug("received non-envelope message, ignoring", "err", err)
		return
	}

	switch env.Kind {
	case ito.KindJoinAccepted:
		slog.Debug("join accepted by loom")
	case ito.KindTaskRequest:
		d.handleTask(ctx, env.Payload, reply)
	case ito.KindExecutorRegister:
		d.handleExecutorRegister(ctx, env.Payload, reply)
	default:
		slog.Warn("unknown message kind", "kind", env.Kind)
	}
}

func (d *Dispatcher) handleTask(ctx context.Context, payload json.RawMessage, reply func(protocol.Message)) {
	var req ito.TaskRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		slog.Error("failed to unmarshal task request", "err", err)
		return
	}

	slog.Info("task received", "task_id", req.TaskID, "kind", req.Kind)

	exec, ok := d.executors[req.Kind]
	if !ok {
		slog.Warn("unknown executor kind", "kind", req.Kind)
		d.sendResult(reply, req.TaskID, false, nil, "unknown kind: "+req.Kind)
		return
	}

	output, err := exec.Execute(ctx, req.Params)
	if err != nil {
		slog.Error("executor failed", "task_id", req.TaskID, "kind", req.Kind, "err", err)
		d.sendResult(reply, req.TaskID, false, nil, err.Error())
		return
	}

	outBytes, _ := json.Marshal(output)
	slog.Info("task completed", "task_id", req.TaskID, "kind", req.Kind)
	d.sendResult(reply, req.TaskID, true, outBytes, "")
}

func (d *Dispatcher) handleExecutorRegister(ctx context.Context, payload json.RawMessage, reply func(protocol.Message)) {
	var reg ito.ExecutorRegistration
	if err := json.Unmarshal(payload, &reg); err != nil {
		slog.Error("failed to unmarshal executor registration", "err", err)
		d.sendRegistered(reply, reg.CorrelationID, false, "invalid payload: "+err.Error())
		return
	}

	if err := d.store.Save(ctx, reg); err != nil {
		slog.Error("failed to persist executor", "kind", reg.Kind, "err", err)
		d.sendRegistered(reply, reg.CorrelationID, false, "store error: "+err.Error())
		return
	}

	exec := store.NewScriptExecutor(reg)
	d.manager.Register(exec)
	d.executors[exec.Name()] = exec

	slog.Info("executor registered", "kind", reg.Kind)
	d.sendRegistered(reply, reg.CorrelationID, true, "")
}

func (d *Dispatcher) sendResult(reply func(protocol.Message), taskID string, success bool, output json.RawMessage, errMsg string) {
	data, err := ito.Encode(ito.KindTaskResult, ito.TaskResult{
		TaskID:  taskID,
		Success: success,
		Output:  output,
		Error:   errMsg,
	})
	if err != nil {
		slog.Error("failed to encode task result", "err", err)
		return
	}
	reply(protocol.Message{Type: websocket.TextMessage, Data: data})
}

func (d *Dispatcher) sendRegistered(reply func(protocol.Message), correlationID string, success bool, errMsg string) {
	data, err := ito.Encode(ito.KindExecutorRegistered, ito.ExecutorRegisteredResult{
		CorrelationID: correlationID,
		Success:       success,
		Error:         errMsg,
	})
	if err != nil {
		slog.Error("failed to encode registered response", "err", err)
		return
	}
	reply(protocol.Message{Type: websocket.TextMessage, Data: data})
}
