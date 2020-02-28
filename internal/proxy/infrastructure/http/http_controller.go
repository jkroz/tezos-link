package http

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/octo-technology/tezos-link/backend/config"
	"github.com/octo-technology/tezos-link/backend/internal/proxy/domain/model"
	"github.com/octo-technology/tezos-link/backend/internal/proxy/usecases"
	"github.com/sirupsen/logrus"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/middleware/stdlib"
	"github.com/ulule/limiter/drivers/store/memory"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Controller represent an http proxy controller
type Controller struct {
	router       *chi.Mux
	uc           usecases.ProxyUsecaseInterface
	reverseProxy *httputil.ReverseProxy
	httpServer   *http.Server
	UUIDRegexp   *regexp.Regexp
}

const (
	uuidRegex = `(?m)([0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12})`
)

// NewHTTPController returns a new http controller
func NewHTTPController(uc usecases.ProxyUsecaseInterface, rp *httputil.ReverseProxy, srv *http.Server) *Controller {
	return &Controller{
		uc:           uc,
		reverseProxy: rp,
		httpServer:   srv,
		UUIDRegexp:   regexp.MustCompile(uuidRegex),
	}
}

// Initialize setup the limiter middleware and handle routes
func (p *Controller) Initialize() {
	basePath := "v1/"
	middleware := setupLimiterMiddleware()
	http.Handle("/"+basePath, middleware.Handler(http.HandlerFunc(handleProxying(p, basePath))))
}

// Run runs the controller
func (p *Controller) Run() {
	log.Fatal(p.httpServer.ListenAndServe())
}

func setupLimiterMiddleware() *stdlib.Middleware {
	rate := limiter.Rate{
		Period: time.Duration(config.ProxyConfig.Proxy.RateLimitPeriod) * time.Second,
		Limit:  config.ProxyConfig.Proxy.RateLimitCount,
	}
	store := memory.NewStore()
	middleware := stdlib.NewMiddleware(limiter.New(store, rate), stdlib.WithForwardHeader(true))
	return middleware
}

func handleProxying(p *Controller, basePath string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, receivedRequest *http.Request) {
		var request = model.NewRequest(
			getRPCFromPath(basePath, receivedRequest.URL.Path, p.UUIDRegexp),
			getUUIDFromPath(receivedRequest.URL.Path, p.UUIDRegexp),
			getActionFromHTTPMethod(receivedRequest.Method),
			receivedRequest.RemoteAddr)
		logrus.Debug(request.Action, request.Path, request.UUID, request.RemoteAddr)

		r, toRawProxy, err := p.uc.Proxy(&request)
		if err != nil {
			logrus.Error(fmt.Sprintf("could not proxy request: %s", err))
		}

		if toRawProxy {
			forwardRawRequestAndRespond(p, w, receivedRequest, &request)
			return
		}

		respondToRequest(w, r)
	}
}

func getUUIDFromPath(path string, re *regexp.Regexp) string {
	var rpcPath string
	for _, match := range re.FindAllString(path, -1) {
		rpcPath = match
	}
	return rpcPath
}

func getRPCFromPath(basePath string, path string, re *regexp.Regexp) string {
	return strings.Replace(path, "/"+basePath+getUUIDFromPath(path, re), "", -1)
}

func forwardRawRequestAndRespond(p *Controller, w http.ResponseWriter, receivedRequest *http.Request, request *model.Request) {
	reverseURL, err := url.Parse("http://dummy" + request.Path)
	if err != nil {
		log.Fatal(fmt.Sprintf("could not construct revers URL: %s", err))
	}
	receivedRequest.URL = reverseURL

	p.reverseProxy.ServeHTTP(w, receivedRequest)
}

func respondToRequest(w http.ResponseWriter, r string) {
	optionsHeaders(w)

	if strings.Contains(r, usecases.NoProxyResponse) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
		return
	}

	_, _ = fmt.Fprint(w, r)
}

func optionsHeaders(w http.ResponseWriter) {
	w.Header().Set("Allow", "OPTIONS, PUSH")
	w.Header().Set("Accept", "application/json")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Depth, User-Agent, X-File-Size, X-Requested-With, If-Modified-Since, X-File-Name, Cache-Control")
	w.Header().Set("Access-Control-Allow-Methods", "PUSH")
	w.Header().Set("Content-Type", "application/json")
}

func getActionFromHTTPMethod(action string) model.Action {
	switch action {
	case "GET":
		return model.OBTAIN
	case "POST":
		return model.PUSH
	case "PUT":
		return model.MODIFY
	}

	return -1
}