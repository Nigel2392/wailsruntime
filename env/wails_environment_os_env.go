//go:build !js && !wasm
// +build !js,!wasm

package env

import "os"

func (c *Environment) GetEnv(key string) WailsResponse[string] {
	var v, ok = os.LookupEnv(key)
	return WailsResponse[string]{
		Data: v,
		OK:   ok,
	}
}

func (c *Environment) SetEnv(key, value string) WailsResponse[bool] {
	var err = os.Setenv(key, value)
	if err != nil {
		return WailsResponse[bool]{
			Err: err.Error(),
		}
	}
	return WailsResponse[bool]{
		Data: true,
		OK:   true,
	}
}

func (c *Environment) UnsetEnv(key string) WailsResponse[bool] {
	var err = os.Unsetenv(key)
	if err != nil {
		return WailsResponse[bool]{
			Err: err.Error(),
		}
	}
	return WailsResponse[bool]{
		Data: true,
		OK:   true,
	}
}
