//go:build !js && !wasm
// +build !js,!wasm

package env

import (
	"context"
	"net/http"
)

type Environment struct {
	ctx context.Context
	ipc IPC
}

func NewEnvironment() *Environment {
	return &Environment{
		ipc: NewIPC(base_ipc_url),
	}
}

func (c *Environment) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.ipc.ServeHTTP(w, r)
}

func (c *Environment) RegisterCallback(name string, fn any) {
	c.ipc.RegisterCallback(name, fn)
}

func (c *Environment) StartUp(ctx context.Context) {
	c.ctx = ctx
}

func (c *Environment) Shutdown(ctx context.Context) {

}
