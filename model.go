package goPool

type ReqPayload interface{}
type RetPayload interface{}

type Worker interface {
	Process(payload ReqPayload) RetPayload
}

type InitializeWorker struct {
	processor func(payload ReqPayload) RetPayload
}

func (w *InitializeWorker) Process(payload ReqPayload) RetPayload {
	return w.processor(payload)
}

type RequestChannel struct {
	workerId int
	worker   Worker
	input    ReqPayload
	retChan  chan RetPayload
}
