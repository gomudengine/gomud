package util

import "sync"

// Hook is a generic, type-safe synchronous callback chain. Handlers registered
// against a Hook receive the current value, may modify it, and return the
// (possibly modified) value. All registered handlers are called in registration
// order. The final value is returned to the caller.
//
// Hook is safe for concurrent registration and firing.
//
// Example – defining a hook point in a package:
//
//	var OnRoomLook util.Hook[RoomTemplateDetails]
//
// Example – registering a handler from a module:
//
//	rooms.OnRoomLook.Register(func(d rooms.RoomTemplateDetails) rooms.RoomTemplateDetails {
//	    d.RoomAlerts = append(d.RoomAlerts, "You can fish here!")
//	    return d
//	})
//
// Example – firing the hook at the call site:
//
//	details = rooms.OnRoomLook.Fire(details)
type Hook[T any] struct {
	mu       sync.RWMutex
	handlers []func(T) T
}

// Register adds a handler to the hook. Handlers are called in the order they
// are registered.
func (h *Hook[T]) Register(fn func(T) T) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers = append(h.handlers, fn)
}

// Fire calls each registered handler in order, passing the return value of
// each as the input to the next. Returns the final value. If no handlers are
// registered the original value is returned unchanged.
func (h *Hook[T]) Fire(data T) T {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, fn := range h.handlers {
		data = fn(data)
	}
	return data
}
