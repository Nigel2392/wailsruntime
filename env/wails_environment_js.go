package env

import (
	"syscall/js"

	"github.com/Nigel2392/jsext/v2/jsc"
	"github.com/Nigel2392/wailsruntime"
)

func init() {
	jsc.BASE64 = true
}

func GetEnv(key string) (data string, ok bool) {
	var resp = environWailsCall[string]("GetEnv", true, key)
	return resp.Data, resp.OK
}

func SetEnv(key, value string) error {
	var resp = environWailsCall[bool]("SetEnv", true, key, value)
	return resp.AsError()
}

func UnsetEnv(key string) error {
	var resp = environWailsCall[bool]("UnsetEnv", true, key)
	return resp.AsError()
}

func OpenFile(constraint *FileConstraint) (File, error) {
	var wailsResp = environWailsCall[File]("OpenFile", true, constraint)
	return wailsResp.Data, wailsResp.AsError()
}

func OpenMultipleFiles(constraint *MultipleFileConstraint) ([]File, error) {
	var wailsResp = environWailsCall[[]File]("OpenMultipleFiles", true, constraint)
	return wailsResp.Data, wailsResp.AsError()
}

func SaveFile(file File, flags FileFlags) error {
	var resp = environWailsCall[bool]("SaveFile", true, file, flags)
	return resp.AsError()
}

func environWailsCall[T any](funcName string, needsRetArgs bool, args ...any) WailsResponse[T] {
	var err error
	if args, err = jsc.ValuesOfInterface(args...); err != nil {
		return WailsResponse[T]{
			Err: err.Error(),
		}
	}

	var wailsRespChan = make(chan WailsResponse[T], 1)
	defer close(wailsRespChan)

	wailsruntime.WailsCall("env", "Environment", funcName, func(args []js.Value) any {
		if len(args) == 0 && needsRetArgs {

			wailsRespChan <- WailsResponse[T]{
				Err: "no arguments passed to callback",
			}

			return nil
		} else if len(args) == 0 {

			wailsRespChan <- WailsResponse[T]{
				OK: true,
			}

			return nil
		}

		if len(args) > 1 {
			var wailsResponse = new(WailsResponse[T])
			if err := jsc.Scan(args[0], wailsResponse); err != nil {
				wailsResponse.Err = err.Error()
				wailsRespChan <- *wailsResponse
				return nil
			}
			wailsRespChan <- *wailsResponse
			return nil
		}

		wailsRespChan <- WailsResponse[T]{
			OK: true,
		}
		return nil
	}, args...)

	var wailsResp = <-wailsRespChan

	return wailsResp
}
