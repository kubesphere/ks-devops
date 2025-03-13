/*
Copyright 2019 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apiserver

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	rt "runtime"
	"time"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/kubesphere/ks-devops/assets"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha1"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	devopsbearertoken "github.com/kubesphere/ks-devops/pkg/apiserver/authentication/authenticators/bearertoken"
	"github.com/kubesphere/ks-devops/pkg/apiserver/authentication/request/anonymous"
	"github.com/kubesphere/ks-devops/pkg/apiserver/filters"
	"github.com/kubesphere/ks-devops/pkg/apiserver/request"
	"github.com/kubesphere/ks-devops/pkg/apiserver/swagger"
	"github.com/kubesphere/ks-devops/pkg/client/cache"
	"github.com/kubesphere/ks-devops/pkg/client/devops"
	"github.com/kubesphere/ks-devops/pkg/client/k8s"
	"github.com/kubesphere/ks-devops/pkg/client/s3"
	"github.com/kubesphere/ks-devops/pkg/client/sonarqube"
	apiserverconfig "github.com/kubesphere/ks-devops/pkg/config"
	"github.com/kubesphere/ks-devops/pkg/indexers"
	"github.com/kubesphere/ks-devops/pkg/informers"
	"github.com/kubesphere/ks-devops/pkg/kapis/common"
	devopsv1alpha2 "github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha2"
	devopsv1alpha3 "github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha3"
	gitops "github.com/kubesphere/ks-devops/pkg/kapis/gitops/v1alpha1"
	"github.com/kubesphere/ks-devops/pkg/kapis/oauth"
	"github.com/kubesphere/ks-devops/pkg/kapis/proxy"
	"github.com/kubesphere/ks-devops/pkg/models/auth"
	utilnet "github.com/kubesphere/ks-devops/pkg/utils/net"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/request/bearertoken"
	unionauth "k8s.io/apiserver/pkg/authentication/request/union"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog/v2"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type APIServer struct {
	// number of kubesphere apiserver
	ServerCount int

	Server *http.Server

	Config *apiserverconfig.Config

	// webservice container, where all webservice defines
	container *restful.Container

	// kubeClient is a collection of all kubernetes(include CRDs) objects clientset
	KubernetesClient k8s.Client

	// informerFactory is a collection of all kubernetes(include CRDs) objects informers,
	// mainly for fast query
	InformerFactory informers.InformerFactory

	// cache is used for short lived objects, like session
	CacheClient cache.Interface

	DevopsClient devops.Interface

	S3Client s3.Interface

	SonarClient sonarqube.SonarInterface

	// controller-runtime cache
	RuntimeCache runtimecache.Cache

	Client client.Client
}

func (s *APIServer) PrepareRun(stopCh <-chan struct{}) error {
	s.container = restful.NewContainer()
	s.container.Filter(logRequestAndResponse)
	s.container.Router(restful.CurlyRouter{})
	// reference: https://pkg.go.dev/github.com/emicklei/go-restful#hdr-Performance_options
	s.container.DoNotRecover(false)
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})

	s.InstallDevOpsAPIs()
	s.setProxy()

	swaggerConfig := swagger.GetSwaggerConfig(s.container)
	s.container.Add(restfulspec.NewOpenAPIService(swaggerConfig))
	s.container.Handle("/swagger-ui/", http.FileServer(http.FS(assets.Static)))

	for _, ws := range s.container.RegisteredWebServices() {
		klog.Infof("Register %s", ws.RootPath())
		for _, r := range ws.Routes() {
			klog.Infof("    %s", r.Method+" "+r.Path)
		}
	}

	s.Server.Handler = s.container

	s.buildHandlerChain(stopCh)

	return nil
}

// Install all DevOps api groups
// Installation happens before all informers start to cache objects,
// so any attempt to list objects using listers will get empty results.
func (s *APIServer) InstallDevOpsAPIs() {
	jenkinsCore := core.JenkinsCore{
		URL:      s.Config.JenkinsOptions.Host,
		UserName: s.Config.JenkinsOptions.Username,
		Token:    s.Config.JenkinsOptions.ApiToken,
	}

	_, err := devopsv1alpha2.AddToContainer(s.container,
		s.InformerFactory.KubeSphereSharedInformerFactory(),
		s.DevopsClient,
		s.SonarClient,
		s.KubernetesClient.KubeSphere(),
		s.S3Client,
		s.Config.JenkinsOptions.Host,
		s.KubernetesClient,
		jenkinsCore)
	utilruntime.Must(err)
	devopsv1alpha3.AddToContainer(s.container, s.DevopsClient, s.KubernetesClient, s.Client, s.RuntimeCache, jenkinsCore, s.Config)
	oauth.AddToContainer(s.container,
		auth.NewTokenOperator(
			s.CacheClient,
			s.Config.AuthenticationOptions),
	)
	gitops.AddToContainer(s.container, &common.Options{
		GenericClient: s.Client,
	}, s.Config.ArgoCDOption, s.Config.FluxCDOption)
}

func (s *APIServer) setProxy() {
	proxy.AddToContainer(s.container)
}

func (s *APIServer) SetContainer(container *restful.Container) {
	s.container = container
}

func (s *APIServer) Run(stopCh context.Context) (err error) {
	if err := indexers.CreatePipelineRunSCMRefNameIndexer(s.RuntimeCache); err != nil {
		return err
	}
	if err := indexers.CreatePipelineRunIdentityIndexer(s.RuntimeCache); err != nil {
		return err
	}

	err = s.waitForResourceSync(stopCh)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-stopCh.Done()
		_ = s.Server.Shutdown(ctx)
	}()

	klog.V(0).Infof("Start listening on %s", s.Server.Addr)
	klog.V(0).Infof("Open the swagger-ui from http://localhost%s/apidocs/?url=http://localhost:9090/apidocs.json", s.Server.Addr)
	if s.Server.TLSConfig != nil {
		err = s.Server.ListenAndServeTLS("", "")
	} else {
		err = s.Server.ListenAndServe()
	}

	return err
}

func (s *APIServer) buildHandlerChain(stopCh <-chan struct{}) {
	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes:          sets.NewString("api", "apis", "kapis", "kapi"),
		GrouplessAPIPrefixes: sets.NewString("api", "kapi"),
	}

	handler := s.Server.Handler
	handler = filters.WithKubeAPIServer(handler, s.KubernetesClient.Config(), &errorResponder{})

	authenticators := make([]authenticator.Request, 0)
	authenticators = append(authenticators, anonymous.NewAuthenticator())

	switch s.Config.AuthMode {
	case apiserverconfig.AuthModeToken:
		authenticators = append(authenticators, bearertoken.New(devopsbearertoken.New()))
	default:
		// TODO error handle
	}

	handler = filters.WithAuthentication(handler, unionauth.New(authenticators...))
	handler = filters.WithRequestInfo(handler, requestInfoResolver)

	s.Server.Handler = handler
}

func (s *APIServer) waitForResourceSync(stopCh context.Context) error {
	klog.V(0).Info("Start cache objects")

	discoveryClient := s.KubernetesClient.Kubernetes().Discovery()
	_, apiResourcesList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return err
	}

	isResourceExists := func(resource schema.GroupVersionResource) bool {
		for _, apiResource := range apiResourcesList {
			if apiResource.GroupVersion == resource.GroupVersion().String() {
				for _, rsc := range apiResource.APIResources {
					if rsc.Name == resource.Resource {
						return true
					}
				}
			}
		}
		return false
	}

	s.InformerFactory.KubernetesSharedInformerFactory().Start(stopCh.Done())
	s.InformerFactory.KubernetesSharedInformerFactory().WaitForCacheSync(stopCh.Done())

	ksInformerFactory := s.InformerFactory.KubeSphereSharedInformerFactory()

	devopsGVRs := []schema.GroupVersionResource{
		{Group: v1alpha1.GroupVersion.Group, Version: v1alpha1.GroupVersion.Version, Resource: "s2ibinaries"},
		{Group: v1alpha1.GroupVersion.Group, Version: v1alpha1.GroupVersion.Version, Resource: "s2ibuildertemplates"},
		{Group: v1alpha1.GroupVersion.Group, Version: v1alpha1.GroupVersion.Version, Resource: "s2iruns"},
		{Group: v1alpha1.GroupVersion.Group, Version: v1alpha1.GroupVersion.Version, Resource: "s2ibuilders"},
		{Group: v1alpha3.GroupVersion.Group, Version: v1alpha3.GroupVersion.Version, Resource: "devopsprojects"},
		{Group: v1alpha3.GroupVersion.Group, Version: v1alpha3.GroupVersion.Version, Resource: "pipelines"},
	}

	for _, gvr := range devopsGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err = ksInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}

	ksInformerFactory.Start(stopCh.Done())
	ksInformerFactory.WaitForCacheSync(stopCh.Done())

	// controller runtime cache for resources
	go func() {
		_ = s.RuntimeCache.Start(stopCh)
	}()
	s.RuntimeCache.WaitForCacheSync(stopCh)

	klog.V(0).Info("Finished caching objects")
	return nil
}

func logStackOnRecover(panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i += 1 {
		_, file, line, ok := rt.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	klog.Errorln(buffer.String())

	headers := http.Header{}
	if ct := w.Header().Get("Content-Type"); len(ct) > 0 {
		headers.Set("Accept", ct)
	}

	w.WriteHeader(http.StatusInternalServerError)
	// ignore this error explicitly
	_, _ = w.Write([]byte("Internal server error"))
}

func logRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	// Always log error response
	logWithVerbose := klog.V(4)
	if resp.StatusCode() > http.StatusBadRequest {
		logWithVerbose = klog.V(0)
	}

	logWithVerbose.Infof("%s - \"%s %s %s\" %d %d %dms",
		utilnet.GetRequestIP(req.Request),
		req.Request.Method,
		req.Request.RequestURI,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)
}

type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
	responsewriters.InternalError(w, req, err)
}
