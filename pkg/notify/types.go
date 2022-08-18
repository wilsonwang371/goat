package notify

type Notifier interface {
	SetSubject(string) error
	SetRecipients([]string) error
	SetContent(string) error
	Send() error
	Level() int
	FeatureFlags() uint64
}
