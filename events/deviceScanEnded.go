package events

var DeviceScanStopped deviceScanStopped

type deviceScanStopped struct {
	handlers []interface{ Handle() }
}

func (sr *deviceScanStopped) Register(handler interface{ Handle() }) {
	sr.handlers = append(sr.handlers, handler)
}

func (sr deviceScanStopped) Trigger() {
	for _, handler := range sr.handlers {
		go handler.Handle()
	}
}
