package resources

import (
	"errors"
	"reflect"
	"runtime"
)

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

const (
	Hello                     = "hello"
	Goodbye                   = "goodbye"
	TestRoute                 = "/test/route"
	TestRoutePathParams       = "/test/route/{param}"
	TestRoutePathParamHello   = "/test/route/hello"
	TestRoutePathParamGoodbye = "/test/route/goodbye"
	HelloWorld                = "hello world"
)

var (
	TestError = errors.New("this is a test Waggy Error")
)