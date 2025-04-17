package worker

import "sync"

type Result struct {
	TraceCode string
	Error     error
	Data      interface{}
}

var ResultMap = sync.Map{}
