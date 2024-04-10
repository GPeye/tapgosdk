package events

var TappedData tappedData

type tappedData struct {
	handlers []interface{ Handle(uint8) }
}

func (sr *tappedData) Register(handler interface{ Handle(uint8) }) {
	sr.handlers = append(sr.handlers, handler)
}

func (sr tappedData) Trigger(payload uint8) {
	for _, handler := range sr.handlers {
		go handler.Handle(payload)
	}
}
