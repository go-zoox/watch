//go:build unix

package process

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(wd) == "process" {
		return filepath.Clean(filepath.Join(wd, ".."))
	}
	return wd
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func waitDialTCP(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr, 40*time.Millisecond)
		if err == nil {
			c.Close()
			return nil
		}
		last = err
		time.Sleep(10 * time.Millisecond)
	}
	if last != nil {
		return fmt.Errorf("dial %s: %w (after %s)", addr, last, timeout)
	}
	return fmt.Errorf("dial %s: timeout after %s", addr, timeout)
}

func countLines(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	if len(b) == 0 {
		return 0, nil
	}
	return len(strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")), nil
}

func waitMinLines(t *testing.T, path string, want int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		n, err := countLines(path)
		if err != nil {
			t.Fatalf("read marker: %v", err)
		}
		if n >= want {
			return
		}
		time.Sleep(15 * time.Millisecond)
	}
	n, _ := countLines(path)
	t.Fatalf("want at least %d lines in %s after %s, got %d", want, path, timeout, n)
}

// longSleep keeps the shell alive until SIGKILL; keep modest so leaked
// subprocesses do not pile up if a test fails.
const longSleep = "sleep 120"

// TestRestartRebindsSameTCPPort checks kill+Wait releases the listen port.
// Skipped under -short (spawns subprocess + go build).
func TestRestartRebindsSameTCPPort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TCP integration in short mode")
	}

	root := moduleRoot(t)
	port := freeTCPPort(t)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	errLog := filepath.Join(t.TempDir(), "listen.stderr")

	bin := filepath.Join(t.TempDir(), "listen-bin")
	build := exec.Command("go", "build", "-o", bin, "./testdata/listen")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build test listener: %v\n%s", err, out)
	}

	cmd := fmt.Sprintf(`%q %d 2>%s`, bin, port, errLog)
	p := New(cmd, &Options{Context: root})
	defer func() { _ = p.Stop() }()
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	if err := waitDialTCP(addr, 20*time.Second); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if err := p.Restart(); err != nil {
			t.Fatalf("restart %d: %v", i, err)
		}
		if err := waitDialTCP(addr, 20*time.Second); err != nil {
			if b, rerr := os.ReadFile(errLog); rerr == nil && len(b) > 0 {
				t.Fatalf("%v\nstderr:\n%s", err, b)
			}
			t.Fatal(err)
		}
	}
}

// TestSequentialRestartAppendsLines checks repeated Restart runs the command again.
func TestSequentialRestartAppendsLines(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker")
	shell := fmt.Sprintf(`echo run >> %q && `+longSleep, marker)

	p := New(shell)
	defer func() { _ = p.Stop() }()
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	waitMinLines(t, marker, 1, 8*time.Second)

	// Restart() only waits until the signal is queued; wait for each kill+run
	// to finish by polling the marker file so the test does not time out while
	// workers are still draining.
	for i := 0; i < 3; i++ {
		if err := p.Restart(); err != nil {
			t.Fatalf("restart %d: %v", i, err)
		}
		want := i + 2 // 1 initial line + one echo per completed restart
		waitMinLines(t, marker, want, 12*time.Second)
	}
}

func TestKillWhenNoChildIsSafe(t *testing.T) {
	p := New("true").(*process)
	if err := p.kill(); err != nil {
		t.Fatal(err)
	}
}
