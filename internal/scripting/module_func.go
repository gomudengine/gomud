package scripting

import (
	"github.com/dop251/goja"
)

var (
	moduleFunctions = map[string]map[string]any{}
)

func AddModlueFunction(namespace string, name string, funcRef any) {
	if _, ok := moduleFunctions[namespace]; !ok {
		moduleFunctions[namespace] = map[string]any{}
	}
	moduleFunctions[namespace][name] = funcRef
}

func setModuleFunctions(vm *goja.Runtime) {
	vm.Set("modules", moduleFunctions)
}
