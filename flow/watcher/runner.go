package watcher

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func (m *Manager) runner() {
	var cmd *exec.Cmd
	for {
		<-m.Restart
		if m.PostChangeCommand == "" {
			m.Logger.Print("`post_change_command` not provided in flow.yml")
			return
		}
		if cmd != nil {
			pid := cmd.Process.Pid
			m.Logger.Success("Stopping: PID %d", pid)
			cmd.Process.Kill()
		}

		cmd = exec.Command(m.PostChangeCommand, m.PostChangeArgs...)

		go func() {
			err := m.runAndListen(cmd)
			if err != nil {
				m.Logger.Error(err)
			}
		}()
	}
}

func (m *Manager) runAndListen(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	mw := io.MultiWriter(&stderr, os.Stderr)
	cmd.Stderr = mw
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		m.Logger.Error(fmt.Errorf("%s\n%s", err, stderr.String()))
		return err
	}

	m.Logger.Success("Running: %s (PID: %d)", strings.Join(cmd.Args, " "), cmd.Process.Pid)
	err = cmd.Wait()
	if err != nil {
		if err.Error() == "signal: killed" {
			m.Logger.Success(err)
			return nil
		}
		m.Logger.Error(fmt.Errorf("%s\n%s", err, stderr.String()))
		return err
	}
	return nil

}
