package discovery

type Client interface {
	Connect(options ...interface{}) error
	Services() *Remotes
	NotificationsChn() chan struct{}
}
