//go:build integration

// Package integration holds the project's integration and end-to-end tests,
// kept separate from the packages under test. They are excluded from the
// default build by the `integration` tag; run them with `make test-integration`
// (or `go test -tags=integration ./...`).
//
// Two harnesses live here:
//   - newSQLiteDB / eachDialect build a real, migrated database in-process for
//     storage-level integration tests.
//   - startServer builds and runs the actual server binary, then drives it over
//     real HTTP for black-box end-to-end tests.
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/pressly/goose/v3"

	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
	"github.com/lennardclaproth/my-finances-tracker/migrations"
)

// ---------------------------------------------------------------------------
// One-time build of the server binary, shared by all end-to-end tests.
// ---------------------------------------------------------------------------

var (
	buildDir  string
	buildOnce sync.Once
	builtBin  string
	buildErr  error
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "mft-itest-bin")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create build dir: %v\n", err)
		os.Exit(1)
	}
	buildDir = dir
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// serverBinary builds the application binary once per test run and returns its
// path. The build is lazy, so storage-only runs never pay for it.
func serverBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		name := "mft-server"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		builtBin = filepath.Join(buildDir, name)
		cmd := exec.Command("go", "build", "-o", builtBin, "github.com/lennardclaproth/my-finances-tracker/cmd/my-finances-tracker")
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = fmt.Errorf("build server binary: %v\n%s", err, out)
		}
	})
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	return builtBin
}

// ---------------------------------------------------------------------------
// Spawned server (black-box end-to-end).
// ---------------------------------------------------------------------------

type testServer struct {
	baseURL string
	log     *safeBuffer
}

// startServer launches the real server binary against a throwaway SQLite
// database and waits until it is healthy. The process is killed on cleanup.
func startServer(t *testing.T) *testServer {
	t.Helper()
	bin := serverBinary(t)

	workDir := t.TempDir() // holds config.yaml; becomes the server's working dir
	dataDir := t.TempDir() // holds the SQLite DB and disk storage
	dbPath := filepath.Join(dataDir, "app.db")
	port := freePort(t)
	writeConfig(t, workDir, port, dbPath, dataDir)

	logBuf := &safeBuffer{}
	cmd := exec.Command(bin)
	cmd.Dir = workDir
	// Keep the APM agent inert: there is no APM server in tests.
	cmd.Env = append(os.Environ(), "ELASTIC_APM_ACTIVE=false")
	cmd.Stdout = logBuf
	cmd.Stderr = logBuf
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	// Registered after t.TempDir cleanups, so it runs first (LIFO): the process
	// is killed and reaped before the temp dirs are removed.
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	ts := &testServer{
		baseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
		log:     logBuf,
	}
	ts.waitForHealth(t)
	return ts
}

func (ts *testServer) waitForHealth(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(ts.baseURL + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("server did not become healthy within 30s\n--- server log ---\n%s", ts.log.String())
}

func (ts *testServer) get(t *testing.T, path string) (int, []byte) {
	t.Helper()
	resp, err := http.Get(ts.baseURL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body of GET %s: %v", path, err)
	}
	return resp.StatusCode, body
}

func (ts *testServer) postJSON(t *testing.T, path string, payload any) (int, []byte) {
	t.Helper()
	buf, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	resp, err := http.Post(ts.baseURL+path, "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body of POST %s: %v", path, err)
	}
	return resp.StatusCode, body
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve free port: %v", err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

func writeConfig(t *testing.T, dir string, port int, dbPath, dataDir string) {
	t.Helper()
	const tmpl = `server:
  environment: test
  port: %d
database:
  type: sqlite3
  connection_string: "%s"
logging:
  level: error
apm:
  server_url: "http://localhost:8200"
  service_name: "my-finances-tracker-itest"
  environment: test
  log_level: error
  transaction_sample_rate: 0
disk_storage:
  base_path: "%s"
providers:
  marketstack:
    base_uri: ""
  alphavantage:
    base_uri: ""
`
	content := fmt.Sprintf(tmpl, port, filepath.ToSlash(dbPath), filepath.ToSlash(dataDir))
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}
}

// safeBuffer is a goroutine-safe buffer: os/exec writes to it from a copy
// goroutine while the test may read it on failure.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// ---------------------------------------------------------------------------
// In-process database (storage-level integration).
// ---------------------------------------------------------------------------

// newSQLiteDB opens a throwaway SQLite database in a temp directory, applies the
// real embedded migrations, and registers cleanup. It uses the same
// storage.NewDB constructor as production and the same embedded migration SQL
// (migrations.GetFS), run through a quiet goose provider.
func newSQLiteDB(t *testing.T) *storage.DB {
	t.Helper()

	// A temp file, not ":memory:": modernc/sqlite gives each pooled connection
	// its own in-memory database, so migrations applied on one connection would
	// be invisible to the next. t.TempDir is removed for us after the test.
	dsn := filepath.Join(t.TempDir(), "test.db")
	db := storage.NewDB(dsn, storage.Sqlite)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test db: %v", err)
		}
	})

	provider, err := goose.NewProvider(goose.DialectSQLite3, db.DB.DB, migrations.GetFS(storage.Sqlite))
	if err != nil {
		t.Fatalf("create migration provider: %v", err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	return db
}

// eachDialect runs fn as a subtest against every database dialect the
// environment can provide. SQLite always runs (in-process, no infra). Postgres
// is the planned second dialect (see docs/TESTING.md); when added, register a
// t.Run("postgres", ...) here and these tests gain coverage with no change to
// their bodies.
func eachDialect(t *testing.T, fn func(t *testing.T, db *storage.DB)) {
	t.Helper()
	t.Run("sqlite", func(t *testing.T) { fn(t, newSQLiteDB(t)) })
}
