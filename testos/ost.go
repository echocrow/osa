package testos

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	osaPkg "github.com/echocrow/osa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertOrgOS tests OSA implementations that utilize the original OS package.
func AssertOrgOS(t *testing.T, osa osaPkg.I) {
	mkTempDir := t.TempDir

	assertExit := func(t *testing.T) {
		tests := []int{0, 1, 34}
		encodeArgs := func(code int) string { return fmt.Sprint(code) }
		exit := func(env string) {
			code, err := strconv.Atoi(env)
			require.NoError(t, err)
			osa.Exit(code)
			t.Fatal("shoul have exited")
		}
		testOsExitCodes(t, tests, encodeArgs, exit)
	}

	AssertOsa(t, osa, mkTempDir, assertExit)

	assertOsExternals(t, osa)
}

// assertOsExternals tests special external OS-related OSA operations. These
// tests leaverage the os package directly and should thus only be run on OSA
// implementations that utilise os package operations.
func assertOsExternals(t *testing.T, osa osaPkg.I) {
	t.Run("OsGetwd", func(t *testing.T) {
		want, err := os.Getwd()
		require.NoError(t, err)
		got, err := osa.Getwd()
		assert.Equal(t, got, want)
		assert.NoError(t, err)
	})
	t.Run("OsUserCacheDir", func(t *testing.T) {
		want, err := os.UserCacheDir()
		require.NoError(t, err)
		got, err := osa.UserCacheDir()
		assert.Equal(t, got, want)
		assert.NoError(t, err)
	})
	t.Run("OsUserHomeDir", func(t *testing.T) {
		want, err := os.UserHomeDir()
		require.NoError(t, err)
		got, err := osa.UserHomeDir()
		assert.Equal(t, got, want)
		assert.NoError(t, err)
	})
}

// testOsExitCode tests that a program exists with the desired exit code.
// Assessment is done by spawning a new process. This can be relatively slow,
// so running repetitive calls as parallel tests is recommended.
// Due to an implementation detail, any required function arguments must be
// encoded as a single string. The encoding and decoding is left to the caller
// and exitFunc.
func testOsExitCode(
	t *testing.T,
	want int,
	args string,
	exitFunc func(args string),
) {
	if args, ok := os.LookupEnv(testExitEnv); ok {
		exitFunc(args)
		t.Fatal("shoul have exited")
		return
	}
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("%s=%s", testExitEnv, args),
	)
	cmd.Run()
	got := cmd.ProcessState.ExitCode()
	assert.Equal(t, want, got)
}

var testExitEnv = "OSA_TEST_EXIT"

// testOsExitCodes runs a series of exit code tests, asserting that a program
// exists with the desired exit code.
// Due to an implementation detail, any required function arguments must be
// encoded as a single string. The encoding and decoding is left to getArgs
// and exitFunc.
func testOsExitCodes(
	t *testing.T,
	codes []int,
	getArgs func(code int) string,
	exitFunc func(args string),
) {
	for i, code := range codes {
		code := code
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			testOsExitCode(t, code, getArgs(code), exitFunc)
		})
	}
}
