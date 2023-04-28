package rabbit

type SigHandler func(sender any, params ...any)

type Signals struct {
	sigHandlers map[string][]SigHandler
}

var sig *Signals

func init() {
	Sig()
}

func Sig() *Signals {
	if sig == nil {
		sig = NewSignals()
	}
	return sig
}

func NewSignals() *Signals {
	return &Signals{
		sigHandlers: make(map[string][]SigHandler),
	}
}

func (s *Signals) Connect(event string, handler SigHandler) {
	s.sigHandlers[event] = append(s.sigHandlers[event], handler)
}

func (s *Signals) DisConnect(event string) {
	delete(s.sigHandlers, event)
}

func (s *Signals) Emit(event string, sender any, params ...any) {
	handlers, ok := s.sigHandlers[event]
	if !ok {
		return
	}
	for _, handler := range handlers {
		handler(sender, params...)
	}
}
