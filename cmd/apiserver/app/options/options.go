/*
Copyright 2020 KubeSphere Authors

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

package options

import (
	"crypto/tls"
	"flag"
	"fmt"
	"kubesphere.io/devops/pkg/client/cache"
	"kubesphere.io/devops/pkg/client/sonarqube"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"kubesphere.io/devops/pkg/apis"
	"kubesphere.io/devops/pkg/apiserver"
	"kubesphere.io/devops/pkg/client/clientset/versioned/scheme"
	apiserverconfig "kubesphere.io/devops/pkg/config"
	"kubesphere.io/devops/pkg/informers"
	genericoptions "kubesphere.io/devops/pkg/server/options"

	"net/http"
	"strings"

	"kubesphere.io/devops/pkg/client/devops/jenkins"
	"kubesphere.io/devops/pkg/client/k8s"
	"kubesphere.io/devops/pkg/client/s3"
	fakes3 "kubesphere.io/devops/pkg/client/s3/fake"
)

type ServerRunOptions struct {
	ConfigFile              string
	GenericServerRunOptions *genericoptions.ServerRunOptions
	*apiserverconfig.Config

	DebugMode bool
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		Config:                  apiserverconfig.New(),
	}

	return s
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "Don't enable this if you don't know what it means.")
	s.GenericServerRunOptions.AddFlags(fs, s.GenericServerRunOptions)
	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.JenkinsOptions.AddFlags(fss.FlagSet("devops"), s.JenkinsOptions)
	s.SonarQubeOptions.AddFlags(fss.FlagSet("sonarqube"), s.SonarQubeOptions)
	s.S3Options.AddFlags(fss.FlagSet("s3"), s.S3Options)

	fs = fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return fss
}

const fakeInterface string = "FAKE"

// NewAPIServer creates an APIServer instance using given options
func (s *ServerRunOptions) NewAPIServer(stopCh <-chan struct{}) (*apiserver.APIServer, error) {
	apiServer := &apiserver.APIServer{
		Config: s.Config,
	}

	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		return nil, err
	}
	apiServer.KubernetesClient = kubernetesClient

	informerFactory := informers.NewInformerFactories(kubernetesClient.Kubernetes(), kubernetesClient.KubeSphere(),
		kubernetesClient.ApiExtensions())
	apiServer.InformerFactory = informerFactory

	if s.S3Options.Endpoint != "" {
		if s.S3Options.Endpoint == fakeInterface && s.DebugMode {
			apiServer.S3Client = fakes3.NewFakeS3()
		} else {
			s3Client, err := s3.NewS3Client(s.S3Options)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to s3, please check s3 service status, error: %v", err)
			}
			apiServer.S3Client = s3Client
		}
	}

	if s.JenkinsOptions.Host != "" {
		devopsClient, err := jenkins.NewDevopsClient(s.JenkinsOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to jenkins, please check jenkins status, error: %v", err)
		}
		apiServer.DevopsClient = devopsClient
	}

	if s.SonarQubeOptions.Host != "" {
		sonarClient, err := sonarqube.NewSonarQubeClient(s.SonarQubeOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connecto to sonarqube, please check sonarqube status, error: %v", err)
		}
		apiServer.SonarClient = sonarqube.NewSonar(sonarClient.SonarQube())
	}

	var cacheClient cache.Interface
	if s.RedisOptions != nil && len(s.RedisOptions.Host) != 0 {
		if s.RedisOptions.Host == fakeInterface && s.DebugMode {
			apiServer.CacheClient = cache.NewSimpleCache()
		} else {
			cacheClient, err = cache.NewRedisClient(s.RedisOptions, stopCh)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to redis service, please check redis status, error: %v", err)
			}
			apiServer.CacheClient = cacheClient
		}
	} else {
		klog.Warning("ks-apiserver starts without redis provided, it will use in memory cache. " +
			"This may cause inconsistencies when running ks-apiserver with multiple replicas.")
		apiServer.CacheClient = cache.NewSimpleCache()
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.GenericServerRunOptions.InsecurePort),
	}

	if s.GenericServerRunOptions.SecurePort != 0 {
		certificate, err := tls.LoadX509KeyPair(s.GenericServerRunOptions.TlsCertFile, s.GenericServerRunOptions.TlsPrivateKey)
		if err != nil {
			return nil, err
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}
		server.Addr = fmt.Sprintf(":%d", s.GenericServerRunOptions.SecurePort)
	}

	sch := scheme.Scheme
	if err := apis.AddToScheme(sch); err != nil {
		klog.Fatalf("unable add APIs to scheme: %v", err)
	}

	// we create a manager for getting client and cache, although the manager is for creating controller. At last, we
	// won't start it up.
	m, err := manager.New(kubernetesClient.Config(), manager.Options{
		Scheme: sch,
		// disable metrics server needed by controller only
		MetricsBindAddress: "0",
	})
	if err != nil {
		klog.Errorf("unable to create manager for getting client and cache, err = %v", err)
		return nil, err
	}
	apiServer.Client = m.GetClient()
	apiServer.RuntimeCache = m.GetCache()
	apiServer.Server = server
	return apiServer, nil
}
