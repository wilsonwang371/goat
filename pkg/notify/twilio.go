package notify

type twilioNotifier struct {
}

// Send implements Notifier
func (*twilioNotifier) Send() error {
	panic("unimplemented")
}

// SetContent implements Notifier
func (*twilioNotifier) SetContent(string) error {
	panic("unimplemented")
}

// SetSubject implements Notifier
func (*twilioNotifier) SetSubject(string) error {
	panic("unimplemented")
}

// SetRecipient implements Notifier
func (*twilioNotifier) SetRecipients([]string) error {
	panic("unimplemented")
}

func NewTwilioNotifier() Notifier {
	return &twilioNotifier{}
}
