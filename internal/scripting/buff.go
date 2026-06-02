package scripting

import (
	"fmt"
	"os"
	"time"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/colorpatterns"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

var (
	buffVMCache       = make(map[int]*VMWrapper)
	scriptBuffTimeout = 50 * time.Millisecond
)

func ClearBuffVMs() {
	clear(buffVMCache)
}

// PruneBuffVMs is intentionally a no-op. Buff VMs are keyed by buff spec ID
// and are not tied to any instance lifecycle.
func PruneBuffVMs(instanceIds ...int) {
}

// InvalidateBuffVM removes the cached VM for a buff spec so the next call
// reloads the script from disk. Call this after saving a buff script via the
// admin API.
func InvalidateBuffVM(buffId int) {
	delete(buffVMCache, buffId)
}

func TryBuffScriptEvent(eventName string, userId int, mobInstanceId int, buffId int) (bool, error) {

	vmw, err := getBuffVM(buffId)
	if err != nil {
		return false, err
	}

	actorInfo := GetActor(userId, mobInstanceId)
	buffTriggersLeft := actorInfo.characterRecord.Buffs.TriggersLeft(buffId)

	timestart := time.Now()
	defer func() {
		mudlog.Debug("TryBuffScriptEvent()", "eventName", eventName, "buffId", buffId, "time", time.Since(timestart))
	}()
	if onCommandFunc, ok := vmw.GetFunction(eventName); ok {

		// Set forced ansi tag wrappers
		userTextWrap.Set(`buff-text`, ``, `cyan`, colorpatterns.Stretch)
		roomTextWrap.Set(`buff-text`, ``, `cyan`, colorpatterns.Stretch)

		res, err := runCallable(vmw, scriptBuffTimeout, onCommandFunc,
			vmw.VM.ToValue(actorInfo),
			vmw.VM.ToValue(buffTriggersLeft),
		)

		// Reset forced ansi tag wrappers
		userTextWrap.Reset()
		roomTextWrap.Reset()

		if err != nil {
			return false, fmt.Errorf("%s(): %w", eventName, err)
		}

		if boolVal, ok := res.Export().(bool); ok {
			return boolVal, nil
		}
	}

	return false, ErrEventNotFound
}

func TryBuffCommand(cmd string, rest string, userId int, mobInstanceId int, buffId int) (bool, error) {

	vmw, err := getBuffVM(buffId)
	if err != nil {
		return false, err
	}

	sActor := GetActor(userId, mobInstanceId)
	sRoom := GetRoom(sActor.GetRoomId())

	timestart := time.Now()
	defer func() {
		mudlog.Debug("TryBuffCommand()", "cmd", cmd, "buffId", buffId, "time", time.Since(timestart))
	}()

	if onCommandFunc, ok := vmw.GetFunction(`onCommand_` + cmd); ok {

		res, err := runCallable(vmw, scriptBuffTimeout, onCommandFunc,
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(sActor),
			vmw.VM.ToValue(sRoom),
		)
		if err != nil {
			return false, fmt.Errorf("onCommand_%s(): %w", cmd, err)
		}

		if boolVal, ok := res.Export().(bool); ok {
			return boolVal, nil
		}

	} else if onCommandFunc, ok := vmw.GetFunction(`onCommand`); ok {

		res, err := runCallable(vmw, scriptBuffTimeout, onCommandFunc,
			vmw.VM.ToValue(cmd),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(sActor),
			vmw.VM.ToValue(sRoom),
		)
		if err != nil {
			return false, fmt.Errorf("onCommand(): %w", err)
		}

		if boolVal, ok := res.Export().(bool); ok {
			return boolVal, nil
		}
	}

	return false, ErrEventNotFound
}

func getBuffVM(buffId int) (*VMWrapper, error) {

	if vmw, ok := buffVMCache[buffId]; ok {
		if vmw == nil {
			return nil, errNoScript
		}
		if scriptHotReload {
			bSpec := buffs.GetBuffSpec(buffId)
			if bSpec != nil {
				if info, err := os.Stat(bSpec.GetScriptPath()); err == nil {
					if info.ModTime().After(vmw.loadedAt) {
						delete(buffVMCache, buffId)
						// fall through to reload
					} else {
						return vmw, nil
					}
				} else {
					return vmw, nil
				}
			} else {
				return vmw, nil
			}
		} else {
			return vmw, nil
		}
	}

	bSpec := buffs.GetBuffSpec(buffId)
	if bSpec == nil {
		return nil, fmt.Errorf("buff spec not found: %T", bSpec)
	}

	script := bSpec.GetScript()
	if len(script) == 0 {
		buffVMCache[buffId] = nil
		return nil, errNoScript
	}

	vmw, err := loadVM(fmt.Sprintf(`buff-%d`, buffId), script, nil)
	if err != nil {
		return nil, err
	}

	buffVMCache[buffId] = vmw
	return vmw, nil
}
