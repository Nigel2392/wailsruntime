//go:build js && wasm
// +build js,wasm

package wailsruntime

import (
	"syscall/js"
)

// Wails functions

var (
	WindowRuntime js.Value = js.Global().Get("window").Get("runtime")
	WindowGo      js.Value = js.Global().Get("window").Get("go")
)

// Call a wails-runtime bound function.
//
// Due to limitations in javascript, we need to await the promise, which is the reason for the callback.
// ( this could also be fixed if wails exposes an API for their IPC! ;) )
//
// This is best used for go functions which perform longer operations, since the javascript runtime will not be blocked.
// (http request for example)
func WailsCall(pkgName, structName, funcName string, cb func(args []js.Value) any, args ...any) js.Value {
	var structure = getStructure(pkgName, structName)
	// Function is a promise, so we need to call the callback when it resolves
	if cb == nil {
		return structure.Call(funcName, args...)
	}
	var function = structure.Get(funcName)
	if !function.Truthy() {
		panic("function not found: " + funcName)
	}

	return function.Invoke(args...).Call("then", releaseableCallback(cb))
}

// Call and await a wails-runtime bound function.
//
// This is best used for go functions which perform shorter operations, since the javascript runtime will be blocked.
func AwaitWailsCall(pkgName, structName, funcName string, args ...any) []js.Value {
	var structure = getStructure(pkgName, structName)
	// Function is a promise, so we need to call the callback when it resolves
	var function = structure.Get(funcName)
	if !function.Truthy() {
		panic("function not found: " + funcName)
	}
	var retChan = make(chan []js.Value)
	var funcOf = js.FuncOf(func(this js.Value, args []js.Value) any {
		retChan <- args
		return nil
	})
	function.Invoke(args...).Call("then", funcOf)
	var ret = <-retChan
	funcOf.Release()
	return ret
}

// Call a wails-runtime bound function in the main package.
func MainCall(strctName, funcName string, cb func(args []js.Value) any, args ...any) js.Value {
	return WailsCall("main", strctName, funcName, cb, args...)
}

// Register a function to be called from javascript.
func EventsOn(eventName string, callback func(args []js.Value) any) {
	WindowRuntime.Call("EventsOn", eventName, releaseableCallback(callback))
}

func EventsOff(eventName string) {
	WindowRuntime.Call("EventsOff", eventName)
}

func EventsOnce(eventName string, callback func(args []js.Value) any) {
	WindowRuntime.Call("EventsOnce", eventName, releaseableCallback(callback))
}

func EventsOnMultiple(eventNames string, maxCnt int, callback func(args []js.Value) any) {
	WindowRuntime.Call("EventsOnMultiple", eventNames, releaseableCallback(callback), maxCnt)
}

func EventsEmit(eventName string, data any) {
	WindowRuntime.Call("EventEmit", eventName, data)
}

func WindowSetTitle(title string) {
	WindowRuntime.Call("WindowSetTitle", title)
}

func WindowFullscreen() {
	WindowRuntime.Call("WindowFullscreen")
}

func WindowUnFullScreen() {
	WindowRuntime.Call("WindowUnfullscreen")
}

func WindowIsFullScreen() bool {
	return boolCall("WindowIsFullscreen")
}

func WindowCenter() {
	WindowRuntime.Call("WindowCenter")
}

func WindowReload() {
	WindowRuntime.Call("WindowReload")
}

func WindowReloadApp() {
	WindowRuntime.Call("WindowReloadApp")
}

func WindowSetSystemDefaultTheme() {
	WindowRuntime.Call("WindowSetSystemDefaultTheme")
}

func WindowSetLightTheme() {
	WindowRuntime.Call("WindowSetLightTheme")
}

func WindowSetDarkTheme() {
	WindowRuntime.Call("WindowSetDarkTheme")
}

func WindowShow() {
	WindowRuntime.Call("WindowShow")
}

func WindowHide() {
	WindowRuntime.Call("WindowHide")
}

func WindowisNormal() bool {
	return boolCall("WindowIsNormal")
}

type size struct {
	Width  int
	Height int
}

func WindowGetSize() (width, height int) {
	var w, h = sizeCall("WindowGetSize", "w", "h")
	return w, h
}

func WindowSetMinSize(width, height int) {
	WindowRuntime.Call("WindowSetMinSize", width, height)
}

func WindowSetMaxSize(width, height int) {
	WindowRuntime.Call("WindowSetMaxSize", width, height)
}

func WindowSetAlwaysOnTop(flag bool) {
	WindowRuntime.Call("WindowSetAlwaysOnTop", flag)
}

func WindowSetPosition(x, y int) {
	WindowRuntime.Call("WindowSetPosition", x, y)
}

func WindowGetPosition() (x, y int) {
	return sizeCall("WindowGetPosition", "x", "y")
}

func WindowMaximise() {
	WindowRuntime.Call("WindowMaximise")
}

func WindowUnmaximise() {
	WindowRuntime.Call("WindowUnmaximise")
}

func WindowIsMaximised() bool {
	return boolCall("WindowIsMaximised")
}

func WindowToggleMaximise() {
	WindowRuntime.Call("WindowToggleMaximise")
}

func WindowUnminimise() {
	WindowRuntime.Call("WindowUnminimise")
}

func WindowIsMinimised() bool {
	return boolCall("WindowIsMinimised")
}

func WindowMinimise() {
	WindowRuntime.Call("WindowMinimise")
}

func WindowSetbackgroundColor(r, g, b, a int8) {
	WindowRuntime.Call("WindowSetbackgroundColor", r, g, b, a)
}

func BrowserOpen(url string) {
	WindowRuntime.Call("BrowserOpenURL", url)
}

func ExitApp() {
	WindowRuntime.Call("Quit")
}

func Hide() {
	WindowRuntime.Call("Hide")
}

func Show() {
	WindowRuntime.Call("Show")
}

func ClipboardGetText() string {
	var promise = WindowRuntime.Call("ClipboardGetText")
	var ret = make(chan string)
	promise.Call("then", releaseableCallback(func(args []js.Value) any {
		ret <- args[0].String()
		return nil
	}))
	return <-ret
}

func ClipboardSetText(text string) bool {
	return boolCall("ClipboardSetText", text)
}

type EnvInfo struct {
	BuildType string
	Platform  string
	Arch      string
}

func Environment() EnvInfo {
	var promise = WindowRuntime.Call("Environment")
	var ret = make(chan EnvInfo)
	promise.Call("then", releaseableCallback(func(args []js.Value) any {
		ret <- EnvInfo{
			BuildType: args[0].Get("buildType").String(),
			Platform:  args[0].Get("platform").String(),
			Arch:      args[0].Get("arch").String(),
		}
		return nil
	}))
	return <-ret
}

func getStructure(pkgName, structName string) js.Value {
	var pkg = WindowGo.Get(pkgName)
	if !pkg.Truthy() {
		panic("package not found: " + pkgName)
	}
	var structure = pkg.Get(structName)
	if !structure.Truthy() {
		panic("structure not found: " + structName)
	}
	return structure
}

func releaseableCallback(cb func(args []js.Value) any) js.Func {
	var funcOf js.Func
	funcOf = js.FuncOf(func(_ js.Value, args []js.Value) any {
		cb(args)
		funcOf.Release()
		return nil
	})
	return funcOf
}

func boolCall(name string, args ...any) bool {
	var promise = WindowRuntime.Call("WindowIsMaximised", args...)
	var ret = make(chan bool)
	promise.Call("then", releaseableCallback(func(args []js.Value) any {
		ret <- args[0].Bool()
		return nil
	}))
	return <-ret
}

func sizeCall(name string, arg1, arg2 string) (int, int) {
	var promise = WindowRuntime.Call(name)
	var ret = make(chan size)
	promise.Call("then", releaseableCallback(func(args []js.Value) any {
		ret <- size{
			Width:  args[0].Get(arg1).Int(),
			Height: args[0].Get(arg2).Int(),
		}
		return nil
	}))
	var s = <-ret
	return s.Width, s.Height
}
