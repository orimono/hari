package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/orimono/ito"
)

type ScriptExecutor struct {
	reg ito.ExecutorRegistration
}

func NewScriptExecutor(reg ito.ExecutorRegistration) *ScriptExecutor {
	return &ScriptExecutor{reg: reg}
}

func (e *ScriptExecutor) Name() string { return e.reg.Kind }

func (e *ScriptExecutor) Capability() ito.Capability {
	return ito.Capability{
		Kind:              e.reg.Kind,
		Version:           e.reg.Version,
		Platforms:         e.reg.Platforms,
		Risk:              e.reg.Risk,
		RequiresElevation: e.reg.RequiresElevation,
		Warning:           e.reg.Warning,
	}
}

// Execute runs the script with params injected as PARAM_<KEY> environment variables.
func (e *ScriptExecutor) Execute(ctx context.Context, raw json.RawMessage) (any, error) {
	var params map[string]string
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
	}

	var cmd *exec.Cmd
	switch e.reg.Runtime {
	case "bash":
		cmd = exec.CommandContext(ctx, "bash", "-c", e.reg.Script)
	case "python":
		cmd = exec.CommandContext(ctx, "python3", "-c", e.reg.Script)
	case "powershell":
		cmd = exec.CommandContext(ctx, "powershell", "-Command", e.reg.Script)
	default:
		return nil, fmt.Errorf("unsupported runtime: %s", e.reg.Runtime)
	}

	env := os.Environ()
	for k, v := range params {
		env = append(env, "PARAM_"+strings.ToUpper(k)+"="+v)
	}
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, out)
	}
	return map[string]string{"output": string(out)}, nil
}
