package scripting

import (
	"errors"
	"fmt"
	"time"

	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/dop251/goja"
)

type VMWrapper struct {
	VM            *goja.Runtime
	callableCache map[string]goja.Callable
	cacheSize     int
	// maxCacheSize is the upper bound on cached callable lookups.
	// A value of 0 means unlimited — the cache grows to hold every function
	// name that has been looked up at least once.
	maxCacheSize int
	// loadedAt records when this VM was compiled and run. Used for hot-reload
	// comparisons against the script file's modification time.
	loadedAt time.Time
}

func newVMWrapper(vm *goja.Runtime, cacheSize int) *VMWrapper {
	return &VMWrapper{VM: vm, callableCache: make(map[string]goja.Callable, cacheSize), maxCacheSize: cacheSize}
}

func (vmw *VMWrapper) GetFunction(name string) (goja.Callable, bool) {

	fn, ok := vmw.callableCache[name]

	if ok {
		return fn, fn != nil
	}

	fn, ok = goja.AssertFunction(vmw.VM.Get(name))

	if vmw.maxCacheSize == 0 || vmw.cacheSize < vmw.maxCacheSize {
		vmw.cacheSize++
		vmw.callableCache[name] = fn
	}

	return fn, ok
}

// loadVM compiles and runs script source in a new Goja VM, returning a ready
// VMWrapper. scriptLabel is used in compile error messages (e.g. "room-42").
// afterLoad, if non-nil, is called with the raw runtime after RunProgram
// succeeds; it runs under scriptLoadTimeout and is intended for onLoad() hooks.
func loadVM(scriptLabel string, source string, afterLoad func(*goja.Runtime) error) (*VMWrapper, error) {
	vm := goja.New()
	setAllScriptingFunctions(vm)

	prg, err := goja.Compile(scriptLabel, source, false)
	if err != nil {
		return nil, fmt.Errorf("Compile: %w", err)
	}

	tmr := time.AfterFunc(scriptLoadTimeout, func() {
		vm.Interrupt(errTimeout)
	})
	_, err = vm.RunProgram(prg)
	vm.ClearInterrupt()
	tmr.Stop()
	if err != nil {
		return nil, wrapVMError("RunProgram", err)
	}

	if afterLoad != nil {
		tmr = time.AfterFunc(scriptLoadTimeout, func() {
			vm.Interrupt(errTimeout)
		})
		err = afterLoad(vm)
		vm.ClearInterrupt()
		tmr.Stop()
		if err != nil {
			return nil, wrapVMError("onLoad", err)
		}
	}

	vmw := newVMWrapper(vm, 0)
	vmw.loadedAt = time.Now()
	return vmw, nil
}

// wrapVMError wraps a VM execution error with a context prefix and logs it.
func wrapVMError(context string, err error) error {
	finalErr := fmt.Errorf("%s: %w", context, err)
	if _, ok := finalErr.(*goja.Exception); ok {
		mudlog.Error("JSVM", "exception", finalErr)
	} else if errors.Is(finalErr, errTimeout) {
		mudlog.Error("JSVM", "interrupted", finalErr)
	} else {
		mudlog.Error("JSVM", "error", finalErr)
	}
	return finalErr
}

// runCallable executes fn under timeout and returns the result.
// On error the raw VM error is logged and returned; the caller should wrap it
// with additional context (e.g. the function name) before surfacing to callers.
func runCallable(vmw *VMWrapper, timeout time.Duration, fn goja.Callable, args ...goja.Value) (goja.Value, error) {
	tmr := time.AfterFunc(timeout, func() {
		vmw.VM.Interrupt(errTimeout)
	})
	res, err := fn(goja.Undefined(), args...)
	vmw.VM.ClearInterrupt()
	tmr.Stop()
	if err != nil {
		return nil, wrapVMError("call", err)
	}
	return res, nil
}
