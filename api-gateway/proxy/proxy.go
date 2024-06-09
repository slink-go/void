package proxy

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type ReverseProxy struct {
	serviceResolver resolver.ServiceResolver
	pathProcessor   resolver.PathProcessor
	logger          logging.Logger
}

func CreateReverseProxy() *ReverseProxy {
	return &ReverseProxy{
		logger: logging.GetLogger("reverse-proxy"),
	}
}
func (p *ReverseProxy) WithServiceResolver(serviceResolver resolver.ServiceResolver) *ReverseProxy {
	p.serviceResolver = serviceResolver
	return p
}
func (p *ReverseProxy) WithPathProcessor(pathProcessor resolver.PathProcessor) *ReverseProxy {
	p.pathProcessor = pathProcessor
	return p
}

func (p *ReverseProxy) ResolveTarget(path string) (*url.URL, error) {
	if p.pathProcessor == nil {
		panic("path processor not set")
	}
	if p.serviceResolver == nil {
		panic("service resolver not set")
	}
	target, err := p.pathProcessor.UrlResolve(path, p.serviceResolver)
	if err != nil {
		return nil, err
	}
	resolved, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return resolved, nil
}
func (p *ReverseProxy) Proxy(ctx *gin.Context, address *url.URL) *httputil.ReverseProxy {
	pr := httputil.NewSingleHostReverseProxy(address)
	pr.Director = func(request *http.Request) {
		request.Header = ctx.Request.Header
		request.Host = address.Host
		request.URL.Scheme = address.Scheme
		request.URL.Host = address.Host
		request.URL.Path = address.Path
	}
	pr.ModifyResponse = p.modifyResponseHandle(address)
	pr.ErrorHandler = p.errHandle

	pr.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   1 * time.Second, // TODO: configurable variable
			KeepAlive: 5 * time.Second, // TODO: configurable variable
		}).DialContext,
		TLSHandshakeTimeout: 1 * time.Second, // TODO: configurable variable
	}
	return pr
}
func (p *ReverseProxy) modifyResponseHandle(address *url.URL) func(response *http.Response) error {
	return func(response *http.Response) error {
		if response.StatusCode == http.StatusInternalServerError {
			u, s := readBody(response)
			p.logger.Error("%s ,req %s ,with error %d, body:%s", u.String(), address, response.StatusCode, s)
			response.Body = io.NopCloser(bytes.NewReader([]byte(fmt.Sprintf("%s", u.String()))))
		} else if response.StatusCode > 300 {
			_, s := readBody(response)
			p.logger.Error("req %s ,with error %d, body:%s", address, response.StatusCode, s)
			response.Body = io.NopCloser(bytes.NewReader([]byte(s)))
		}
		return nil
	}
}
func (p *ReverseProxy) errHandle(res http.ResponseWriter, req *http.Request, err error) {
	fmt.Println(err)
}

func readBody(response *http.Response) (uuid.UUID, string) {
	defer response.Body.Close()
	all, _ := io.ReadAll(response.Body)
	u := uuid.New()
	var s string
	if len(all) > 0 {
		s = string(all)
	}
	return u, s
}
