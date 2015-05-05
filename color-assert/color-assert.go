package assert

import (
	"fmt"
	"regexp"
	"runtime"
	"testing"
)

var re = regexp.MustCompile(`^.*jp2tileserver\.(.*)$`)

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

func True(expression bool, message string, t *testing.T) {
	caller := getCallerName(1)
	if !expression {
		failure(caller, message, t)
		return
	}
	success(caller, message, t)
}

func False(exp bool, m string, t *testing.T) {
	True(!exp, m, t)
}

func Equal(expected, actual interface{}, message string, t *testing.T) {
	caller := getCallerName(1)
	if expected != actual {
		failure(caller, fmt.Sprintf("Expected %#v, but got %#v - %s", expected, actual, message), t)
		return
	}
	success(caller, message, t)
}
