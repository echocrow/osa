# OSA – OS abstraction package

Go code (golang) packages that offer easy in-memory testing of code calling [os](https://pkg.go.dev/os) functions without reading/writing real files, or requiring extensive refactoring or dependency injection.

## Contents

- [Features](#features)
- [Packages](#packages)
- [Basic Usage](#basic-usage-tldr)
- [Extended Usage](#extended-usage)
- [API Documentation](#api-documentation)

## Features

- Support for most `os` functions (as of Go 1.17).
- [Monkey-patching](https://en.wikipedia.org/wiki/Monkey_patch) in-memory `os` replacement during testing, allowing for:
  - Fast file I/O tests without causing any real filesystem reads or writes.
  - Simple stdio (stdin/stdou/stderr) testing without needing to call a subprocess.
  - Exit code catching & testing without needing to call a subprocess.
- No extensive rewrites or dependency injections required.
- Common `os` assert/require test helper functions included.

## Packages

The following packages are included:

- `osa`: The main OS abstraction package. It determines which `os` functions are supported and tracks the currently active implementation. Implementing packages simply need to import this package instead of `"os"`, no further changes required.
- `osa/oos`: The standard `osa` implementation. This package simply wraps and calls the default `os` functions of the standard library. This is the default `osa` implementation, so typically code does not need to import or directly interact with this package.
- `osa/vos`: The virtual `osa` implementation. This package mimicks `os` features in-memory, so no real files are created, read, updated, or deleted. The package provides a `Patch()` function to inject this implementation for testing. Only test packages need to know about this.
- `osa/testos`: An OS testing helpers library. This package provides useful helper functions for repetitive `os` calls and assert/require operations during testing, such as `RequireWrite()`, `RequireMkdirAll()`, `AssertNotExists()`, `AssertFileData()`, `GetStdio()`, and more.
- `osa/testosa`: An OSA testing library. This package provides assertions for custom OSA implementations.

## Basic Usage (TLDR)

Use `osa` package instead of `os`:
```diff
import (
-	"os"
+	os "github.com/echocrow/osa"
)
```
Patch & use in-memory implementation via `osa/vos` during testing:
```diff
import (
	"testing"
+	"github.com/echocrow/osa/vos"
)

func TestExample(t *testing.T) {
+	os, reset := vos.Patch()
+	defer reset()
	// ...
}
```

## Extended Usage

### Implementation

In the implementing package, use `osa` instead of the original `os` package via an aliased import:

```go
package example

import (
	// Import "osa" instead of the standard "os" package.
	os "github.com/echocrow/osa"
)

// No other code changes required; use "os" as usual, e.g.:
func Example() string {
	if data, err := os.ReadFile("my-file.go"); err != nil {
		return ""
	}
	return data
}
```

By default, this will simply call the original implementation of the standard `os` package.

### Testing

In the testing package, the virtual OS package (`vos`) can be monkey-patched, replacing the standard `os` implementation with an in-memory version:

```go
package example_test

import (
	"testing"
	"example.com/example"
	"github.com/echocrow/osa"
)

func TestExample(t *testing.T) {
	// Monkey-patch virtual OS at the start of the test.
	_, reset := vos.Patch()
	defer reset()

	// Test like you normally would, e.g.:
	got := example.Example()
	if !got.MatchString("my-data") {
		t.Fatal(`¯\_(ツ)_/¯`)
	}
}
```

If you need to call `os` methods during the test, the patch function of the `vos` package also returns the virtual, patched `os` abstraction:

```go
func TestAnotherExample(t *testing.T) {
	// Monkey-patch and get a virtual OS implementation.
	os, reset := vos.Patch()
	defer reset()

	// Set up test like you normally would. All side-effects happen purely
	// in-memory thanks to vos, e.g.:
	cfgDir, _ := os.UserConfigDir()
	cfgPath := filepath.Join(cfgDir, "my-config")
	os.WriteFile(cfgPath, []byte("my-data"), 0666)

 	// ...
}
```

## API Documentation

- See [OSA on pkg.go.dev](https://pkg.go.dev/github.com/echocrow/osa).
