package limiter

type Limiter interface {
	GetRps() int
}
