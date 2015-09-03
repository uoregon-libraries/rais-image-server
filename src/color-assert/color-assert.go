// Package assert offers some very simple helper methods for testing.  Not
// meant for external use per se, though there's nothing tying this to our
// codebase / project.
//
// Assertion methods (Equal, True, False, etc) expect a `message` string to be
// passed in, which should be a simple explanation that will help you
// understand what went wrong, such as "foo.Bar is 25".  Wordy messages won't
// necessarily help debugging as assert functions should report as much
// information as they can about where an assertion went wrong.
package assert

import (
	"fmt"
	"regexp"
	"runtime"
	"testing"
)

var re = regexp.MustCompile(`^.*jp2tileserver\.(.*)$`)

// Caller represents data used by an assertion to show the file/function/line
// of where an assertion went wrong, rather than using the built-in system
// which would report the "failure" function every time, since all asserts that
// fail eventually find their way in there.
type Caller struct {
	Func     *runtime.Func
	Name     string
	Filename string
	Line     int
}

func getCallerName(skip int) *Caller {
	// Increase skip since they surely don't want *this* function
	pc, file, line, _ := runtime.Caller(skip + 1)
	fn := runtime.FuncForPC(pc)
	return &Caller{
		Func:     fn,
		Name:     re.ReplaceAllString(fn.Name(), "$1"),
		Filename: file,
		Line:     line,
	}
}

func success(caller *Caller, message string, t *testing.T) {
	fmt.Printf("\033[32mok\033[0m        %s(): %s\n", caller.Name, message)
}

func failure(caller *Caller, message string, t *testing.T) {
	fmt.Printf("\033[31;1mnot ok\033[0m    %s(): %s\n", caller.Name, message)
	fmt.Printf("          - %s:%d\n", caller.Filename, caller.Line)
	t.FailNow()
}

// True fails the tests if `expression` isn't the boolean value `true`
func True(expression bool, message string, t *testing.T) {
	caller := getCallerName(1)
	if !expression {
		failure(caller, message, t)
		return
	}
	success(caller, message, t)
}

// False is a convenience method wrapping True and negating the expression
func False(exp bool, m string, t *testing.T) {
	True(!exp, m, t)
}

// Equal verifies that `expected` and `actual` are the same as per "!=" rules.
// This makes it work well for simple types, but more complex types will still
// need specialized checks.
func Equal(expected, actual interface{}, message string, t *testing.T) {
	caller := getCallerName(1)
	if expected != actual {
		failure(caller, fmt.Sprintf("Expected %#v, but got %#v - %s", expected, actual, message), t)
		return
	}
	success(caller, message, t)
}

// IncludesString checks `list` for inclusion of `string`, reporting failure if
// it is not present.
func IncludesString(expected string, list []string, message string, t *testing.T) {
	caller := getCallerName(1)
	for _, s := range list {
		if expected == s {
			success(caller, message, t)
			return
		}
	}

	failure(caller, fmt.Sprintf("Expected %#v to be included in %#v - %s", expected, list, message), t)
}
