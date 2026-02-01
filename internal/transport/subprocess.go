package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const (
	defaultMaxBufferSize     = 1024 * 1024 // 1MB buffer limit
	minimumClaudeCodeVersion = "2.0.0"
	sdkVersion               = "0.1.0"
)

var cmdLengthLimit = func() int {
	if runtime.GOOS == "windows" {
		return 8000
	}
	return 100000
}()

// SubprocessTransport implements Transport using the Claude Code CLI subprocess.
type SubprocessTransport struct {
	prompt        string
	isStreaming   bool
	options       *Options
	cliPath       string
	cwd           string
	process       *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	ready         bool
	exitError     error
	maxBufferSize int
	tempFiles     []string
	writeMu       sync.Mutex
	closeMu       sync.Mutex
	closed        bool
}

// NewSubprocessTransport creates a new subprocess transport.
func NewSubprocessTransport(prompt string, isStreaming bool, options *Options) (*SubprocessTransport, error) {
	t := &SubprocessTransport{
		prompt:        prompt,
		isStreaming:   isStreaming,
		options:       options,
		maxBufferSize: defaultMaxBufferSize,
	}

	if options.MaxBufferSize > 0 {
		t.maxBufferSize = options.MaxBufferSize
	}

	if options.Cwd != "" {
		t.cwd = options.Cwd
	}

	if options.CLIPath != "" {
		t.cliPath = options.CLIPath
	} else {
		path, err := t.findCLI()
		if err != nil {
			return nil, err
		}
		t.cliPath = path
	}

	return t, nil
}

func (t *SubprocessTransport) findCLI() (string, error) {
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	home, _ := os.UserHomeDir()
	locations := []string{
		filepath.Join(home, ".npm-global", "bin", "claude"),
		"/usr/local/bin/claude",
		filepath.Join(home, ".local", "bin", "claude"),
		filepath.Join(home, "node_modules", ".bin", "claude"),
		filepath.Join(home, ".yarn", "bin", "claude"),
		filepath.Join(home, ".claude", "local", "claude"),
	}

	for _, path := range locations {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}

	return "", fmt.Errorf("claude code not found, install with:\n" +
		"  npm install -g @anthropic-ai/claude-code\n\n" +
		"or provide the path via Options:\n" +
		"  WithCLIPath(\"/path/to/claude\")")
}

func (t *SubprocessTransport) buildSettingsValue() (string, error) {
	hasSettings := t.options.Settings != ""
	hasSandbox := t.options.Sandbox != nil

	if !hasSettings && !hasSandbox {
		return "", nil
	}

	if hasSettings && !hasSandbox {
		return t.options.Settings, nil
	}

	settingsObj := make(map[string]any)

	if hasSettings {
		settingsStr := strings.TrimSpace(t.options.Settings)
		if strings.HasPrefix(settingsStr, "{") && strings.HasSuffix(settingsStr, "}") {
			if err := json.Unmarshal([]byte(settingsStr), &settingsObj); err != nil {
				data, err := os.ReadFile(settingsStr)
				if err == nil {
					_ = json.Unmarshal(data, &settingsObj)
				}
			}
		} else {
			data, err := os.ReadFile(settingsStr)
			if err == nil {
				_ = json.Unmarshal(data, &settingsObj)
			}
		}
	}

	if hasSandbox {
		settingsObj["sandbox"] = t.options.Sandbox
	}

	data, err := json.Marshal(settingsObj)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (t *SubprocessTransport) buildCommand() ([]string, error) {
	cmd := []string{t.cliPath, "--output-format", "stream-json", "--verbose"}

	if t.options.SystemPrompt == nil {
		cmd = append(cmd, "--system-prompt", "")
	} else if s, ok := t.options.SystemPrompt.(string); ok {
		cmd = append(cmd, "--system-prompt", s)
	} else if preset, ok := t.options.SystemPrompt.(*SystemPromptPreset); ok {
		if preset.Type == "preset" && preset.Append != "" {
			cmd = append(cmd, "--append-system-prompt", preset.Append)
		}
	}

	if t.options.Tools != nil {
		if tools, ok := t.options.Tools.([]string); ok {
			if len(tools) == 0 {
				cmd = append(cmd, "--tools", "")
			} else {
				cmd = append(cmd, "--tools", strings.Join(tools, ","))
			}
		} else {
			cmd = append(cmd, "--tools", "default")
		}
	}

	if len(t.options.AllowedTools) > 0 {
		cmd = append(cmd, "--allowedTools", strings.Join(t.options.AllowedTools, ","))
	}

	if t.options.MaxTurns > 0 {
		cmd = append(cmd, "--max-turns", strconv.Itoa(t.options.MaxTurns))
	}

	if t.options.MaxBudgetUSD != nil {
		cmd = append(cmd, "--max-budget-usd", fmt.Sprintf("%g", *t.options.MaxBudgetUSD))
	}

	if len(t.options.DisallowedTools) > 0 {
		cmd = append(cmd, "--disallowedTools", strings.Join(t.options.DisallowedTools, ","))
	}

	if t.options.Model != "" {
		cmd = append(cmd, "--model", t.options.Model)
	}

	if t.options.FallbackModel != "" {
		cmd = append(cmd, "--fallback-model", t.options.FallbackModel)
	}

	if len(t.options.Betas) > 0 {
		cmd = append(cmd, "--betas", strings.Join(t.options.Betas, ","))
	}

	if t.options.PermissionPromptToolName != "" {
		cmd = append(cmd, "--permission-prompt-tool", t.options.PermissionPromptToolName)
	}

	if t.options.PermissionMode != "" {
		cmd = append(cmd, "--permission-mode", t.options.PermissionMode)
	}

	if t.options.ContinueConversation {
		cmd = append(cmd, "--continue")
	}

	if t.options.Resume != "" {
		cmd = append(cmd, "--resume", t.options.Resume)
	}

	settingsValue, err := t.buildSettingsValue()
	if err != nil {
		return nil, err
	}
	if settingsValue != "" {
		cmd = append(cmd, "--settings", settingsValue)
	}

	for _, dir := range t.options.AddDirs {
		cmd = append(cmd, "--add-dir", dir)
	}

	if t.options.MCPServers != nil {
		switch servers := t.options.MCPServers.(type) {
		case map[string]any:
			if len(servers) > 0 {
				data, err := json.Marshal(map[string]any{"mcpServers": servers})
				if err != nil {
					return nil, err
				}
				cmd = append(cmd, "--mcp-config", string(data))
			}
		case string:
			cmd = append(cmd, "--mcp-config", servers)
		}
	}

	if t.options.IncludePartialMessages {
		cmd = append(cmd, "--include-partial-messages")
	}

	if t.options.ForkSession {
		cmd = append(cmd, "--fork-session")
	}

	if len(t.options.Agents) > 0 {
		agentsJSON, err := json.Marshal(t.options.Agents)
		if err != nil {
			return nil, err
		}
		cmd = append(cmd, "--agents", string(agentsJSON))
	}

	if t.options.SettingSources != nil {
		cmd = append(cmd, "--setting-sources", strings.Join(t.options.SettingSources, ","))
	} else {
		cmd = append(cmd, "--setting-sources", "")
	}

	for _, plugin := range t.options.Plugins {
		if plugin.Type == "local" {
			cmd = append(cmd, "--plugin-dir", plugin.Path)
		}
	}

	for flag, value := range t.options.ExtraArgs {
		if value == nil {
			cmd = append(cmd, "--"+flag)
		} else {
			cmd = append(cmd, "--"+flag, *value)
		}
	}

	if t.options.MaxThinkingTokens > 0 {
		cmd = append(cmd, "--max-thinking-tokens", strconv.Itoa(t.options.MaxThinkingTokens))
	}

	if t.options.OutputFormat != nil {
		if t.options.OutputFormat["type"] == "json_schema" {
			if schema, ok := t.options.OutputFormat["schema"]; ok {
				schemaJSON, err := json.Marshal(schema)
				if err != nil {
					return nil, err
				}
				cmd = append(cmd, "--json-schema", string(schemaJSON))
			}
		}
	}

	if t.isStreaming {
		cmd = append(cmd, "--input-format", "stream-json")
	} else {
		cmd = append(cmd, "--print", "--", t.prompt)
	}

	cmdStr := strings.Join(cmd, " ")
	if len(cmdStr) > cmdLengthLimit && len(t.options.Agents) > 0 {
		agentsJSON, _ := json.Marshal(t.options.Agents)
		tempFile, err := os.CreateTemp("", "claude-agents-*.json")
		if err != nil {
			return nil, err
		}
		if _, err := tempFile.Write(agentsJSON); err != nil {
			_ = tempFile.Close()
			return nil, err
		}
		_ = tempFile.Close()
		t.tempFiles = append(t.tempFiles, tempFile.Name())

		for i, arg := range cmd {
			if arg == "--agents" && i+1 < len(cmd) {
				cmd[i+1] = "@" + tempFile.Name()
				break
			}
		}
	}

	return cmd, nil
}

func (t *SubprocessTransport) checkClaudeVersion(ctx context.Context) error {
	if os.Getenv("CLAUDE_AGENT_SDK_SKIP_VERSION_CHECK") != "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2e9)
	defer cancel()

	cmd := exec.CommandContext(ctx, t.cliPath, "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	versionOutput := strings.TrimSpace(string(output))
	re := regexp.MustCompile(`([0-9]+\.[0-9]+\.[0-9]+)`)
	match := re.FindString(versionOutput)
	if match == "" {
		return nil
	}

	if compareVersions(match, minimumClaudeCodeVersion) < 0 {
		fmt.Fprintf(os.Stderr, "Warning: Claude Code version %s is unsupported in the Agent SDK. "+
			"Minimum required version is %s. Some features may not work correctly.\n",
			match, minimumClaudeCodeVersion)
	}

	return nil
}

func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		aNum, _ := strconv.Atoi(aParts[i])
		bNum, _ := strconv.Atoi(bParts[i])
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
	}

	if len(aParts) < len(bParts) {
		return -1
	}
	if len(aParts) > len(bParts) {
		return 1
	}
	return 0
}

// Connect starts the subprocess.
func (t *SubprocessTransport) Connect(ctx context.Context) error {
	if t.process != nil {
		return nil
	}

	if err := t.checkClaudeVersion(ctx); err != nil {
		return err
	}

	args, err := t.buildCommand()
	if err != nil {
		return err
	}

	t.process = exec.CommandContext(ctx, args[0], args[1:]...)

	env := os.Environ()
	for k, v := range t.options.Env {
		env = append(env, k+"="+v)
	}
	env = append(env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")
	env = append(env, "CLAUDE_AGENT_SDK_VERSION="+sdkVersion)
	if t.options.EnableFileCheckpointing {
		env = append(env, "CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING=true")
	}
	if t.cwd != "" {
		env = append(env, "PWD="+t.cwd)
	}
	t.process.Env = env

	if t.cwd != "" {
		t.process.Dir = t.cwd
	}

	var stdinErr, stdoutErr, stderrErr error
	t.stdin, stdinErr = t.process.StdinPipe()
	t.stdout, stdoutErr = t.process.StdoutPipe()
	t.stderr, stderrErr = t.process.StderrPipe()

	if stdinErr != nil || stdoutErr != nil || stderrErr != nil {
		return fmt.Errorf("failed to create pipes")
	}

	if err := t.process.Start(); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("claude code not found at: %s", t.cliPath)
		}
		return fmt.Errorf("failed to start claude code: %w", err)
	}

	go t.handleStderr()

	if !t.isStreaming {
		_ = t.stdin.Close()
		t.stdin = nil
	}

	t.ready = true
	return nil
}

func (t *SubprocessTransport) handleStderr() {
	if t.stderr == nil {
		return
	}

	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if t.options.Stderr != nil {
			t.options.Stderr(line)
		} else if t.options.DebugStderr != nil {
			_, _ = fmt.Fprintln(t.options.DebugStderr, line)
		}
	}
}

// Write sends raw data to the transport.
func (t *SubprocessTransport) Write(ctx context.Context, data string) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	if !t.ready || t.stdin == nil {
		return fmt.Errorf("transport is not ready for writing")
	}

	if t.process != nil && t.process.ProcessState != nil && t.process.ProcessState.Exited() {
		return fmt.Errorf("cannot write to terminated process (exit code: %d)",
			t.process.ProcessState.ExitCode())
	}

	if t.exitError != nil {
		return fmt.Errorf("cannot write to process that exited with error: %w", t.exitError)
	}

	if _, err := io.WriteString(t.stdin, data); err != nil {
		t.ready = false
		t.exitError = fmt.Errorf("failed to write to process stdin: %w", err)
		return t.exitError
	}

	return nil
}

// ReadMessages returns a channel that receives parsed JSON messages.
func (t *SubprocessTransport) ReadMessages(ctx context.Context) <-chan ReadResult {
	ch := make(chan ReadResult, 100)

	go func() {
		defer close(ch)

		if t.process == nil || t.stdout == nil {
			ch <- ReadResult{Error: fmt.Errorf("not connected")}
			return
		}

		reader := bufio.NewReader(t.stdout)
		var jsonBuffer strings.Builder

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				ch <- ReadResult{Error: fmt.Errorf("error reading stdout: %w", err)}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			jsonBuffer.WriteString(line)

			if jsonBuffer.Len() > t.maxBufferSize {
				bufLen := jsonBuffer.Len()
				jsonBuffer.Reset()
				ch <- ReadResult{Error: fmt.Errorf("JSON message exceeded maximum buffer size of %d bytes (size: %d)",
					t.maxBufferSize, bufLen)}
				continue
			}

			var data map[string]any
			if err := json.Unmarshal([]byte(jsonBuffer.String()), &data); err != nil {
				continue
			}

			jsonBuffer.Reset()
			ch <- ReadResult{Data: data}
		}

		if t.process != nil {
			_ = t.process.Wait()
			if t.process.ProcessState != nil && !t.process.ProcessState.Success() {
				exitCode := t.process.ProcessState.ExitCode()
				if exitCode != 0 {
					t.exitError = fmt.Errorf("command failed with exit code %d", exitCode)
					ch <- ReadResult{Error: t.exitError}
				}
			}
		}
	}()

	return ch
}

// Close closes the transport and cleans up resources.
func (t *SubprocessTransport) Close() error {
	t.closeMu.Lock()
	defer t.closeMu.Unlock()

	if t.closed {
		return nil
	}
	t.closed = true

	for _, f := range t.tempFiles {
		_ = os.Remove(f)
	}
	t.tempFiles = nil

	t.ready = false

	t.writeMu.Lock()
	if t.stdin != nil {
		_ = t.stdin.Close()
		t.stdin = nil
	}
	t.writeMu.Unlock()

	if t.stderr != nil {
		_ = t.stderr.Close()
		t.stderr = nil
	}

	if t.process != nil && t.process.Process != nil {
		_ = t.process.Process.Kill()
		_ = t.process.Wait()
	}

	t.process = nil
	t.stdout = nil
	t.exitError = nil

	return nil
}

// IsReady returns true if the transport is ready for communication.
func (t *SubprocessTransport) IsReady() bool {
	return t.ready
}

// EndInput ends the input stream (closes stdin).
func (t *SubprocessTransport) EndInput() error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	if t.stdin != nil {
		err := t.stdin.Close()
		t.stdin = nil
		return err
	}
	return nil
}
