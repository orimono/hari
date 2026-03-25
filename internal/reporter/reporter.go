package reporter

type Sender interface {
	Send([]byte) error
}

type Reporter struct {
	sender Sender
}

func NewReporter(sender Sender) *Reporter {
	return &Reporter{
		sender: sender,
	}
}

func (r *Reporter) Send(data []byte) {
	r.sender.Send(data)
}
