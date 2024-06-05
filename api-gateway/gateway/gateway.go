package gateway

type Gateway interface {
	Serve(addresses ...string)
}
