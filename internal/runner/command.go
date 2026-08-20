package runner

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

// ResolveAuthorization determines if execution is authorized based on ADR-0007.
func ResolveAuthorization(flagSet bool, envVal string, interactivePrompt func(string) (string, error)) bool {
	if flagSet || envVal == "1" {
		return true
	}
	if interactivePrompt != nil {
		resp, err := interactivePrompt("execute all command bindings in the invocation?")
		if err == nil && resp == "yes" {
			return true
		}
	}
	return false
}

type limitWriter struct {
	buf   *bytes.Buffer
	limit int
	kill  func()
	killC chan struct{}
}

func (w *limitWriter) Write(p []byte) (int, error) {
	if w.buf.Len()+len(p) > w.limit {
		w.buf.Write(p[:w.limit-w.buf.Len()])
		if w.kill != nil {
			w.kill()
		}
		return len(p), nil
	}
	w.buf.Write(p)
	return len(p), nil
}

func prepareExecTarget(report *VerdictReport, argv []string, planRoot string) (string, *Verdict) {
	if len(argv) == 0 {
		return "", NewBlockedVerdict(report, ReasonArgvInvalid)
	}
	execTarget := argv[0]
	if !strings.Contains(execTarget, "/") {
		return execTarget, nil
	}
	resolved, err := confine(planRoot, execTarget)
	if err != nil {
		return "", NewOperationalVerdict(report, ReasonPathEscapesRoot)
	}
	return resolved, nil
}

func setupCommandEnv(cmd *exec.Cmd, planRoot string) {
	env := os.Environ()
	var newEnv []string
	for _, kv := range env {
		if !strings.HasPrefix(kv, "PWD=") {
			newEnv = append(newEnv, kv)
		}
	}
	newEnv = append(newEnv, "PWD="+planRoot)
	cmd.Env = newEnv
}

func checkStartError(report *VerdictReport, startErr error) *Verdict {
	if startErr == nil {
		return nil
	}
	if errors.Is(startErr, exec.ErrNotFound) || errors.Is(startErr, fs.ErrNotExist) {
		return NewBlockedVerdict(report, ReasonExecutableMissing)
	}
	if errors.Is(startErr, fs.ErrPermission) {
		return NewBlockedVerdict(report, ReasonExecutableDenied)
	}
	return NewOperationalVerdict(report, ReasonProcessStart)
}

func executeCommand(
	report *VerdictReport,
	cmd *exec.Cmd,
	resolvedTimeout time.Duration,
	limitC <-chan struct{},
) (timeoutFired, limitFired bool, errVerdict *Verdict) {
	if startErr := cmd.Start(); startErr != nil {
		return false, false, checkStartError(report, startErr)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	timer := time.NewTimer(resolvedTimeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		timeoutFired = true
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) //nolint:errcheck // kill may fail if already dead
		<-done
	case <-limitC:
		limitFired = true
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) //nolint:errcheck // kill may fail if already dead
		<-done
	case <-done:
	}

	return timeoutFired, limitFired, nil
}

func resolveTimeout(binding *hcr.ExecutableBinding) (resolved time.Duration, declared *time.Duration) {
	timeoutSecs := 120
	if binding.TimeoutSecs > 0 {
		timeoutSecs = binding.TimeoutSecs
		d := time.Duration(timeoutSecs) * time.Second
		declared = &d
	}
	return time.Duration(timeoutSecs) * time.Second, declared
}

func prepareCommand(
	execTarget string,
	argv []string,
	planRoot string,
	report *VerdictReport,
) (*exec.Cmd, *os.File, *Verdict) {
	var args []string
	if len(argv) > 1 {
		args = argv[1:]
	}
	cmd := exec.CommandContext(context.Background(), execTarget, args...)
	cmd.Dir = planRoot
	setupCommandEnv(cmd, planRoot)

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return nil, nil, NewOperationalVerdict(report, ReasonProcessStart)
	}
	cmd.Stdin = devNull
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd, devNull, nil
}

func configureOutputLimits(cmd *exec.Cmd) <-chan struct{} {
	killedLimit := make(chan struct{}, 1)
	killOnce := func() {
		select {
		case killedLimit <- struct{}{}:
		default:
		}
	}
	const maxOutput = 1048576
	cmd.Stdout = &limitWriter{buf: new(bytes.Buffer), limit: maxOutput, kill: killOnce, killC: killedLimit}
	cmd.Stderr = &limitWriter{buf: new(bytes.Buffer), limit: maxOutput, kill: killOnce, killC: killedLimit}
	return killedLimit
}

func parseCommandBinding(binding *hcr.ExecutableBinding, report *VerdictReport) ([]string, *Verdict) {
	if binding.Pattern != "" || binding.Select != "" {
		return nil, NewBlockedVerdict(report, ReasonModifierUnimplemented)
	}
	argv := strings.Fields(binding.Run)
	if len(argv) == 0 {
		return nil, NewBlockedVerdict(report, ReasonArgvInvalid)
	}
	return argv, nil
}

// EvaluateCommandBinding executes a command binding under the ADR-0008 contract.
func EvaluateCommandBinding(
	authorized bool,
	target TargetProvenance,
	binding *hcr.ExecutableBinding,
	planRoot string,
) *Verdict {
	report := VerdictReport{
		Identity: BindingIdentity{RecordPath: binding.RecordPath, BindingIndex: binding.BindingIndex},
		Target:   target, Kind: binding.Kind, Severity: binding.Severity,
	}

	argv, errVerdict := parseCommandBinding(binding, &report)
	if errVerdict != nil {
		return errVerdict
	}

	execTarget, errVerdict := prepareExecTarget(&report, argv, planRoot)
	if errVerdict != nil {
		return errVerdict
	}

	if !authorized {
		return NewBlockedVerdict(&report, ReasonAuthorizationDenied)
	}

	resolvedTimeout, declared := resolveTimeout(binding)
	report.Timeouts = TimeoutRecord{Declared: declared, Resolved: resolvedTimeout}

	cmd, devNull, errVerdict := prepareCommand(execTarget, argv, planRoot, &report)
	if errVerdict != nil || cmd == nil {
		return errVerdict
	}
	defer devNull.Close()

	limitC := configureOutputLimits(cmd)

	start := time.Now()
	timeoutFired, limitFired, errVerdict := executeCommand(&report, cmd, resolvedTimeout, limitC)
	if errVerdict != nil {
		return errVerdict
	}

	report.Execution = &Execution{
		Argv:     argv,
		ExitCode: cmd.ProcessState.ExitCode(),
		Duration: time.Since(start),
	}

	if limitFired {
		return NewOperationalVerdict(&report, ReasonLimitExceeded)
	}
	if timeoutFired {
		return NewBlockedVerdict(&report, ReasonTimeout)
	}
	if !cmd.ProcessState.Success() {
		report.Findings = append(report.Findings, Finding{Severity: binding.Severity})
	}
	return NewCompletedVerdict(&report)
}
