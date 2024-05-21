package discovery

type Client interface {
	Connect() error
	Services() *Remotes
}
