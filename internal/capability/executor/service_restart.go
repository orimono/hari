package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/orimono/ito"
	cap "github.com/orimono/shutter/internal/capability"
)

type ServiceRestartExecutor struct{}

func (e *ServiceRestartExecutor) Name() string { return "service.restart" }

func (e *ServiceRestartExecutor) Capability() ito.Capability {
	return ito.Capability{
		Kind:              "service.restart",
		Version:           "1.0.0",
		Platforms:         []string{"linux", "darwin", "windows"},
		Risk:              ito.RiskMedium,
		RequiresElevation: true,
		Warning:           "重启服务期间相关业务将短暂中断。",
	}
}

type serviceRestartParams struct {
	Name string `json:"name"`
}

func (p *serviceRestartParams) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func (e *ServiceRestartExecutor) Execute(ctx context.Context, raw json.RawMessage) (any, error) {
	params, err := cap.ParseParams[serviceRestartParams](raw)
	if err != nil {
		return nil, err
	}
	name := params.Name

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "systemctl", "restart", name)
	case "darwin":
		cmd = exec.CommandContext(ctx, "launchctl", "kickstart", "-k", "system/"+name)
	case "windows":
		cmd = exec.CommandContext(ctx, "powershell", "-Command", "Restart-Service", name)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, out)
	}
	return map[string]string{"output": string(out)}, nil
}
