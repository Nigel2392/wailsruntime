//go:build !js && !wasm
// +build !js,!wasm

package env

import (
	"context"
)

type Environment struct {
	ctx context.Context
}

func NewEnvironment() *Environment {
	return &Environment{}
}

func (c *Environment) StartUp(ctx context.Context) {
	c.ctx = ctx
}

func (c *Environment) Shutdown(ctx context.Context) {

}
