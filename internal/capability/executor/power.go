package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/orimono/ito"
)

type ShutdownExecutor struct{}

func (e *ShutdownExecutor) Name() string { return "power.shutdown" }

func (e *ShutdownExecutor) Capability() ito.Capability {
	return ito.Capability{
		Kind:              "power.shutdown",
		Version:           "1.0.0",
		Platforms:         []string{"linux", "darwin", "windows"},
		Risk:              ito.RiskHigh,
		RequiresElevation: true,
		Warning:           "节点将立即关机，需手动或通过 WoL 才能重新上线。",
	}
}

func (e *ShutdownExecutor) Execute(ctx context.Context, _ json.RawMessage) (any, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.CommandContext(ctx, "shutdown", "-h", "now")
	case "windows":
		cmd = exec.CommandContext(ctx, "shutdown", "/s", "/t", "0")
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, out)
	}
	return map[string]string{"output": string(out)}, nil
}

type RebootExecutor struct{}

func (e *RebootExecutor) Name() string { return "power.reboot" }

func (e *RebootExecutor) Capability() ito.Capability {
	return ito.Capability{
		Kind:              "power.reboot",
		Version:           "1.0.0",
		Platforms:         []string{"linux", "darwin", "windows"},
		Risk:              ito.RiskHigh,
		RequiresElevation: true,
		Warning:           "节点将立即重启，重启期间所有服务不可用。",
	}
}

func (e *RebootExecutor) Execute(ctx context.Context, _ json.RawMessage) (any, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.CommandContext(ctx, "shutdown", "-r", "now")
	case "windows":
		cmd = exec.CommandContext(ctx, "shutdown", "/r", "/t", "0")
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, out)
	}
	return map[string]string{"output": string(out)}, nil
}
