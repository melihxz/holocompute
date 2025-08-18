package holocompute

import (
	"github.com/melihxz/holocompute/pkg/proto"
)

// TaskSpec specifies a task to be executed
type TaskSpec struct {
	// Module is the WASM module to execute
	Module WASMModule

	// Func is the function to call in the module
	Func string

	// Inputs are the input arrays
	Inputs Inputs

	// Outputs are the output arrays
	Outputs Outputs

	// ResourceHints provides hints about resource requirements
	ResourceHints ResourceHints
}

// WASMModule represents a WASM module
type WASMModule struct {
	// Bytes contains the WASM bytecode
	Bytes []byte

	// SHA256 is the SHA256 hash of the module
	SHA256 []byte
}

// Inputs maps input names to shared arrays
type Inputs map[string]SharedArray

// Outputs maps output names to shared arrays
type Outputs map[string]SharedArray

// ResourceHints provides hints about resource requirements
type ResourceHints struct {
	// CPU is the number of CPU cores required
	CPU int32

	// GPU indicates if a GPU is required
	GPU bool

	// MemoryMB is the amount of memory required in MB
	MemoryMB int32
}

// TaskResult represents the result of a task
type TaskResult struct {
	// Status is the status of the task
	Status TaskStatus

	// Outputs are the output references
	Outputs Outputs

	// Logs contains any logs from the task execution
	Logs string
}

// TaskStatus represents the status of a task
type TaskStatus int

const (
	// TaskPending means the task is waiting to be scheduled
	TaskPending TaskStatus = iota

	// TaskRunning means the task is currently executing
	TaskRunning

	// TaskSuccess means the task completed successfully
	TaskSuccess

	// TaskFailed means the task failed
	TaskFailed

	// TaskTimeout means the task timed out
	TaskTimeout
)

// MustLoadWASM loads a WASM module from a file, panicking on error
func MustLoadWASM(filename string) WASMModule {
	// TODO: Implement WASM loading
	return WASMModule{}
}

// ToProto converts a ResourceHints to a protobuf ResourceHints
func (rh ResourceHints) ToProto() *proto.ResourceHints {
	return &proto.ResourceHints{
		Cpu:      rh.CPU,
		Gpu:      rh.GPU,
		MemoryMb: rh.MemoryMB,
	}
}

// FromProto converts a protobuf ResourceHints to a ResourceHints
func (rh *ResourceHints) FromProto(p *proto.ResourceHints) {
	rh.CPU = p.Cpu
	rh.GPU = p.Gpu
	rh.MemoryMB = p.MemoryMb
}
