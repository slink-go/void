package registry

type ServiceRegistry interface {
	Get(applicationId string) (string, error)
}
