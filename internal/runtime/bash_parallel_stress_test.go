package runtime

import (
	"context"
	"sync"
	"testing"

	"github.com/riipandi/elph/pkg/tools"
)

func TestBashParallelStress(t *testing.T) {
	t.Parallel()
	const n = 32
	wd := t.TempDir()
	var wg sync.WaitGroup
	errs := make(chan string, n)
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := ExecuteTool(context.Background(), wd, tools.Bash, map[string]any{
				"command": "echo hello",
			})
			if result.Err != nil || result.Output != "hello" {
				errs <- result.Output
			}
		}()
	}
	wg.Wait()
	close(errs)
	for bad := range errs {
		t.Fatalf("bad output %q", bad)
	}
}
