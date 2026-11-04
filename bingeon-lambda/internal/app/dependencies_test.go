package app_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// slices are the features and the other slices each may import: only
// account builds on others, to compose the view. A new slice adds its row.
var featureSlices = map[string][]string{
	"catalog":    {},
	"favourites": {},
	"picks":      {},
	"account":    {"favourites", "picks"},
}

const module = "bingeon-lambda/"

// TestDependenciesPointInward checks every non-test import:
//   - platform imports only platform: it knows nothing of the features;
//   - a slice imports platform and what its row allows;
//   - within a slice, domain.go does no I/O (no AWS, Gemini, TMDB or
//     HTTP), only store.go touches DynamoDB, and the service never sees
//     the router;
//   - only app, cmd and tests import app; only tests import testutil.
func TestDependenciesPointInward(t *testing.T) {
	err := filepath.WalkDir("../..", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(filepath.ToSlash(path), "../../")
		for _, imp := range f.Imports {
			dep, _ := strconv.Unquote(imp.Path.Value)
			if why := forbidden(rel, dep); why != "" {
				t.Errorf("%s imports %s: %s", rel, dep, why)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// forbidden says why the file at path (from bingeon-lambda/) may not
// import dep; "" when it may.
func forbidden(path, dep string) string {
	pkg := filepath.ToSlash(filepath.Dir(path))
	internal, isInternal := strings.CutPrefix(dep, module+"internal/")
	file := filepath.Base(path)
	top := func(p string) string { return strings.SplitN(strings.TrimPrefix(p, "internal/"), "/", 2)[0] }

	switch {
	case isInternal && top(internal) == "testutil" && top(pkg) != "testutil":
		return "testutil is for tests"
	case isInternal && top(internal) == "app" && top(pkg) != "app" && !strings.HasPrefix(pkg, "cmd/"):
		return "only the entry points build the app"
	case top(pkg) == "platform" && isInternal && top(internal) != "platform":
		return "platform knows nothing of the features"
	}

	name, isSlice := strings.CutPrefix(pkg, "internal/")
	if _, ok := featureSlices[name]; !isSlice || !ok {
		return ""
	}
	if isInternal && top(internal) != "platform" && !slices.Contains(featureSlices[name], top(internal)) {
		return "not in its row in featureSlices (dependencies_test.go)"
	}
	io := []string{"github.com/aws/", "google.golang.org/genai", "net/http", "os"}
	switch {
	case file == "domain.go" && (slices.ContainsFunc(io, func(p string) bool { return strings.HasPrefix(dep, p) }) ||
		dep == module+"internal/platform/dynamo" || dep == module+"internal/platform/tmdb" || dep == module+"internal/platform/gemini"):
		return "domain.go does no I/O"
	case file != "store.go" && (strings.HasPrefix(dep, "github.com/aws/") || dep == module+"internal/platform/dynamo"):
		return "only store.go touches DynamoDB"
	case file == "service.go" && dep == module+"internal/platform/router":
		return "the service doesn't know how it is called"
	}
	return ""
}

func TestForbiddenCatchesEachRule(t *testing.T) {
	for _, c := range []struct{ path, dep string }{
		{"internal/favourites/domain.go", "bingeon-lambda/internal/platform/dynamo"},
		{"internal/favourites/domain.go", "net/http"},
		{"internal/favourites/service.go", "github.com/aws/aws-sdk-go-v2/service/dynamodb"},
		{"internal/favourites/service.go", "bingeon-lambda/internal/platform/router"},
		{"internal/favourites/handler.go", "bingeon-lambda/internal/picks"},
		{"internal/platform/router/router.go", "bingeon-lambda/internal/catalog"},
		{"internal/picks/store.go", "bingeon-lambda/internal/testutil"},
	} {
		if forbidden(c.path, c.dep) == "" {
			t.Errorf("%s may import %s", c.path, c.dep)
		}
	}
	for _, c := range []struct{ path, dep string }{
		{"internal/favourites/store.go", "bingeon-lambda/internal/platform/dynamo"},
		{"internal/account/domain.go", "bingeon-lambda/internal/picks"},
		{"cmd/api/main.go", "bingeon-lambda/internal/app"},
	} {
		if why := forbidden(c.path, c.dep); why != "" {
			t.Errorf("%s may not import %s: %s", c.path, c.dep, why)
		}
	}
}
