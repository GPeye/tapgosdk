package events

var RawData rawData

type rawData struct {
	handlers []interface{ Handle([]byte) }
}

func (sr *rawData) Register(handler interface{ Handle([]byte) }) {
	sr.handlers = append(sr.handlers, handler)
}

func (sr rawData) Trigger(payload []byte) {
	for _, handler := range sr.handlers {
		go handler.Handle(payload)
	}
}
