package worker

import "sync"

type Result struct {
	TraceCode string
	Error     error
}

var ResultMap = sync.Map{}