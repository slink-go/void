package discovery

type Client interface {
	Connect() error
	Services() *Remotes
}

//func NewDiscoClientConfig() *disco.Config {
//	return disco.NewConfig()
//}
//func NewDiscoClient(config *disco.Config) Client {
//	return disco.NewDiscoClient(config)
//}
//
//func NewEurekaClientConfig() *eureka.Config {
//	return eureka.NewConfig()
//}
//func NewEurekaClient(config *eureka.Config) Client {
//	return eureka.NewEurekaClient(config)
//}
//
//func NewStaticClient(services map[string][]remote.Remote) Client {
//	return static.NewStaticClient(services)
//}
//func NewStaticClientFromFile(path string) (Client, error) {
//	return static.LoadFromFile(path)
//}
