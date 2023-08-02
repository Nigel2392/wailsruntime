//go:build !js && !wasm
// +build !js,!wasm

package env

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/Nigel2392/mux"
)

type callback struct {
	fn      any
	typeOf  reflect.Type
	valueOf reflect.Value
}

func newCallback(fn any) callback {
	var c = callback{
		fn:      fn,
		typeOf:  reflect.TypeOf(fn),
		valueOf: reflect.ValueOf(fn),
	}
	return c
}

func (c callback) Call(args []any) ([]interface{}, error) {
	var argsVals = make([]reflect.Value, 0, len(args))

	if len(args) != c.typeOf.NumIn() {
		return nil, fmt.Errorf("wrong number of args: expected %d, got %d", c.typeOf.NumIn(), len(args))
	}

	for i, arg := range args {

		var argType = reflect.TypeOf(arg)
		var paramType = c.typeOf.In(i)

		if argType != paramType && !argType.ConvertibleTo(paramType) {
			return nil, fmt.Errorf("wrong type for arg %d: expected %s, got %s", i, paramType.String(), argType.String())
		}

		if argType != paramType {
			arg = reflect.ValueOf(arg).Convert(paramType).Interface()
		}

		argsVals = append(argsVals, reflect.ValueOf(arg))
	}

	if len(argsVals) != c.typeOf.NumIn() {
		return nil, fmt.Errorf("wrong number of argsVals: expected %d, got %d, something went wrong setting the fields", c.typeOf.NumIn(), len(argsVals))
	}

	var retVals = c.valueOf.Call(argsVals)
	var ret = make([]interface{}, 0)
	var err error
	for _, retVal := range retVals {
		var iFace = retVal.Interface()
		if ifErr, ok := iFace.(error); ok && err == nil {
			err = ifErr
		} else {
			ret = append(ret, iFace)
		}
	}
	return ret, err
}

type ipc struct {
	mux       *mux.Mux
	callbacks map[string]callback
}

type IPC interface {
	RegisterCallback(name string, fn any)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

var LogFunc func(args []interface{}) error

func NewIPC(basePath string) IPC {
	var i = &ipc{
		mux:       mux.New(),
		callbacks: make(map[string]callback),
	}
	var path = strings.Trim(basePath, "/")
	var environPath = fmt.Sprintf("/%s/<<funcname>>", path)
	i.mux.Handle(mux.ANY, environPath, i.handler())
	i.mux.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if LogFunc != nil {
			LogFunc([]interface{}{"IPC", "404", r.URL.Path})
		}
		errorResponse(w, 500, "page not found")
	})
	return i
}

func (c *ipc) RegisterCallback(name string, fn any) {
	if LogFunc != nil {
		LogFunc([]interface{}{"IPC", "RegisterCallback", name})
	}
	c.callbacks[name] = newCallback(fn)
}

func (c *ipc) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var vars = mux.Vars(r)
		var cb = vars.Get("funcname")
		if cb == "" {
			if LogFunc != nil {
				LogFunc([]interface{}{"IPC", "handler", "no function name specified"})
			}
			errorResponse(w, 500, "no function name specified")
			return
		}

		if !strings.HasSuffix(cb, ".callback") {
			if LogFunc != nil {
				LogFunc([]interface{}{"IPC", "handler", "invalid function name"})
			}
			errorResponse(w, 500, "invalid function name")
			return
		}

		cb = strings.TrimSuffix(cb, ".callback")

		var callback, ok = c.callbacks[cb]
		if !ok {
			if LogFunc != nil {
				LogFunc([]interface{}{"IPC", "handler", "callback not found"})
			}
			errorResponse(w, 500, "callback not found")
			return
		}

		var args any

		if r.Body == nil && callback.typeOf.NumIn() > 0 {
			if LogFunc != nil {
				LogFunc([]interface{}{"IPC", "handler", "no args specified"})
			}
			errorResponse(w, 500, "no args specified")
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			if LogFunc != nil {
				LogFunc([]interface{}{"IPC", "handler", "error decoding args", err.Error()})
			}
			errorResponse(w, 500, err.Error())
			return
		}

		argsSlice, ok := args.([]interface{})
		if !ok {
			if LogFunc != nil {
				LogFunc([]interface{}{"IPC", "handler", "args is not a slice"})
			}
			errorResponse(w, 500, "args is not a slice")
			return
		}

		var resp, err = callback.Call(argsSlice)
		if err != nil {
			if LogFunc != nil {
				LogFunc([]interface{}{"IPC", "handler", "error calling callback", err.Error()})
			}
			errorResponse(w, 500, err.Error())
			return
		}

		if LogFunc != nil {
			LogFunc([]interface{}{"IPC", "handler", "callback called successfully"})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		var wailsResp = WailsResponse[[]interface{}]{
			Data: resp,
			OK:   true,
		}
		json.NewEncoder(w).Encode(wailsResp)
	})
}

func (c *ipc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if LogFunc != nil {
		LogFunc([]interface{}{"IPC", "ServeHTTP", r.URL.Path})
	}
	c.mux.ServeHTTP(w, r)
}

func errorResponse(w http.ResponseWriter, statuscode int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statuscode)
	var m = make(map[string]string)
	m["error"] = err
	json.NewEncoder(w).Encode(m)
}
