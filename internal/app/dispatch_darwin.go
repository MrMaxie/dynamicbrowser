package app

import (
	"context"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/go-webgpu/goffi/ffi"
	"github.com/go-webgpu/goffi/types"
	"github.com/gogpu/systray"
)

var (
	darwinTasks    sync.Map
	darwinTaskID   atomic.Uint64
	darwinCallback = ffi.NewCallback(func(id uintptr) uintptr {
		if task, ok := darwinTasks.LoadAndDelete(id); ok {
			task.(func())()
		}
		return 0
	})
)

func newTrayDispatch(_ *systray.SystemTray, ctx context.Context) (func(func()) error, func(), error) {
	// Share systray's FFI runtime; two zero-CGO runtimes cannot be linked together.
	library, err := ffi.LoadLibrary("/usr/lib/libSystem.B.dylib")
	if err != nil {
		return nil, nil, err
	}
	ready := false
	defer func() {
		if !ready {
			_ = ffi.FreeLibrary(library)
		}
	}()
	mainQueue, err := ffi.GetSymbol(library, "_dispatch_main_q")
	if err != nil {
		return nil, nil, err
	}
	enqueue, err := ffi.GetSymbol(library, "dispatch_async_f")
	if err != nil {
		return nil, nil, err
	}
	var call types.CallInterface
	if err := ffi.PrepareCallInterface(&call, types.DefaultCall, types.VoidTypeDescriptor, []*types.TypeDescriptor{types.PointerTypeDescriptor, types.PointerTypeDescriptor, types.PointerTypeDescriptor}); err != nil {
		return nil, nil, err
	}
	var pending sync.Map
	dispatch := func(job func()) error {
		return dispatchTask(ctx, func(run func()) error {
			id := uintptr(darwinTaskID.Add(1))
			pending.Store(id, struct{}{})
			darwinTasks.Store(id, func() { defer pending.Delete(id); run() })
			// GCD copies these scalar arguments; its context is an ID, never a Go pointer.
			callback := darwinCallback
			_, err := ffi.CallFunction(&call, enqueue, nil, []unsafe.Pointer{unsafe.Pointer(&mainQueue), unsafe.Pointer(&id), unsafe.Pointer(&callback)})
			if err != nil {
				darwinTasks.Delete(id)
				pending.Delete(id)
			}
			return err
		}, job)
	}
	restore := func() {
		pending.Range(func(id, _ any) bool { darwinTasks.Delete(id); return true })
		_ = ffi.FreeLibrary(library)
	}
	ready = true
	return dispatch, restore, nil
}
