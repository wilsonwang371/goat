package notify

type pushoverNotifier struct{}

// Send implements Notifier
func (*pushoverNotifier) Send() error {
	panic("unimplemented")
}

// SetContent implements Notifier
func (*pushoverNotifier) SetContent(string) error {
	panic("unimplemented")
}

// SetSubject implements Notifier
func (*pushoverNotifier) SetSubject(string) error {
	panic("unimplemented")
}

// SetRecipient implements Notifier
func (*pushoverNotifier) SetRecipients([]string) error {
	panic("unimplemented")
}

func NewPushoverNotifier() Notifier {
	return &pushoverNotifier{}
}
