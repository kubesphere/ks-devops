/*

  Copyright 2023 The KubeSphere Authors.

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

package v1alpha1

import (
	"context"
	"strings"

	"github.com/emicklei/go-restful"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"kubesphere.io/devops/pkg/apiserver/query"
	devopsClient "kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/kapis"
	resourcesV1alpha3 "kubesphere.io/devops/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// StrategyMapping Currently only support `buildpacks-v3` strategy
var StrategyMapping = map[string]string{
	"nodejs":  "buildpacks-v3-full",
	"go":      "buildpacks-v3-go",
	"python":  "buildpacks-v3-python",
	"java":    "buildpacks-v3-java",
	"default": "buildpacks-v3-full",
}

// apiHandlerOption holds some useful tools for API handler.
type apiHandlerOption struct {
	devopsClient devopsClient.Interface
	client       client.Client
}

// apiHandler contains functions to handle coming request and give a response.
type apiHandler struct {
	apiHandlerOption
}

// newAPIHandler creates an APIHandler.
func newAPIHandler(o apiHandlerOption) *apiHandler {
	return &apiHandler{o}
}

func (h *apiHandler) listImageBuildStrategies(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)

	strategyList := &buildv1alpha1.ClusterBuildStrategyList{}

	if err := h.client.List(context.Background(), strategyList); err != nil {
		kapis.HandleError(request, response, err)
		return
	}

	apiResult := resourcesV1alpha3.DefaultList(toBuildStrategyObjects(strategyList.Items),
		queryParam,
		resourcesV1alpha3.DefaultCompare(),
		resourcesV1alpha3.DefaultFilter(), nil)

	_ = response.WriteAsJson(apiResult)
}

func toBuildStrategyObjects(apps []buildv1alpha1.ClusterBuildStrategy) []runtime.Object {
	objs := make([]runtime.Object, len(apps))
	for i := range apps {
		objs[i] = &apps[i]
	}
	return objs
}

func (h *apiHandler) getImageBuildStrategy(request *restful.Request, response *restful.Response) {
	strategyName := request.PathParameter("imageBuildStrategy")

	// get imageBuildStrategy
	strategy := &buildv1alpha1.ClusterBuildStrategy{}
	if err := h.client.Get(context.Background(), client.ObjectKey{Name: strategyName}, strategy); err != nil {
		kapis.HandleError(request, response, err)
		return
	}
	_ = response.WriteEntity(strategy)
}

func (h *apiHandler) listImageBuilds(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	queryParam := query.ParseQueryParameter(request)

	opts := make([]client.ListOption, 0, 3)
	opts = append(opts, client.InNamespace(namespace))
	buildList := &buildv1alpha1.BuildList{}

	if err := h.client.List(context.Background(), buildList, opts...); err != nil {
		kapis.HandleError(request, response, err)
		return
	}

	apiResult := resourcesV1alpha3.DefaultList(
		toBuildObjects(buildList.Items),
		queryParam,
		resourcesV1alpha3.DefaultCompare(),
		resourcesV1alpha3.DefaultFilter(), nil)

	_ = response.WriteAsJson(apiResult)
}

func toBuildObjects(apps []buildv1alpha1.Build) []runtime.Object {
	objs := make([]runtime.Object, len(apps))
	for i := range apps {
		objs[i] = &apps[i]
	}
	return objs
}

func (h *apiHandler) createImageBuild(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	imageBuild := request.PathParameter("build")
	sourceUrl := request.QueryParameter("sourceUrl")
	language := request.QueryParameter("language")
	outputImage := request.QueryParameter("outputImage")

	build := buildv1alpha1.Build{}
	err := request.ReadEntity(&build)
	if err != nil {
		klog.Error(err)
		kapis.HandleBadRequest(response, request, err)
		return
	}

	build.Namespace = namespace
	build.Name = imageBuild + "-"
	build.Spec.Source.URL = &sourceUrl

	strategyName, exists := StrategyMapping[strings.ToLower(language)]
	if !exists {
		strategyName = StrategyMapping["default"]
	}

	build.Spec.Strategy.Name = strategyName
	build.Spec.Output.Image = outputImage

	if err := h.client.Create(context.Background(), &build); err != nil {
		kapis.HandleError(request, response, err)
		return
	}

	_ = response.WriteEntity(build)
}

func (h *apiHandler) updateImageBuild(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	imageBuild := request.PathParameter("build")

	oldBuild := buildv1alpha1.Build{}
	if err := h.client.Get(context.Background(), client.ObjectKey{Name: imageBuild}, &oldBuild); err != nil {
		kapis.HandleError(request, response, err)
		return
	}

	sourceUrl := request.QueryParameter("sourceUrl")
	language := request.QueryParameter("language")
	outputImage := request.QueryParameter("outputImage")

	err := request.ReadEntity(&oldBuild)
	if err != nil {
		klog.Error(err)
		kapis.HandleBadRequest(response, request, err)
		return
	}

	oldBuild.Spec.Source.URL = &sourceUrl
	if "nodejs" == language {
		oldBuild.Spec.Strategy.Name = "buildpacks-v3"
	}
	oldBuild.Spec.Output.Image = outputImage
	oldBuild.Namespace = namespace

	if err := h.client.Update(context.Background(), &oldBuild); err != nil {
		kapis.HandleError(request, response, err)
		return
	}

	_ = response.WriteEntity(oldBuild)
}

func (h *apiHandler) getImageBuild(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	imageBuild := request.PathParameter("build")

	// get imageBuild
	build := buildv1alpha1.Build{}
	if err := h.client.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: imageBuild}, &build); err != nil {
		kapis.HandleError(request, response, err)
		return
	}
	_ = response.WriteEntity(&build)
}

func (h *apiHandler) deleteImageBuild(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	imageBuild := request.PathParameter("build")

	// get imageBuild
	build := buildv1alpha1.Build{}
	if err := h.client.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: imageBuild}, &build); err != nil {
		kapis.HandleError(request, response, err)
		return
	}
	if err := h.client.Delete(context.Background(), &build); err != nil {
		kapis.HandleError(request, response, err)
		return
	}
	_ = response.WriteEntity(&build)
}

func (h *apiHandler) createImageBuildRun(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	buildrunName := request.PathParameter("imageBuildrun")
	imageBuild := request.QueryParameter("build")

	buildRun := buildv1alpha1.BuildRun{}
	err := request.ReadEntity(&buildRun)
	if err != nil {
		klog.Error(err)
		kapis.HandleBadRequest(response, request, err)
		return
	}

	buildRun.ObjectMeta.GenerateName = buildrunName + "-"
	buildRun.Spec.BuildRef.Name = imageBuild
	buildRun.Namespace = namespace

	if err := h.client.Create(context.Background(), &buildRun); err != nil {
		kapis.HandleError(request, response, err)
		return
	}

	_ = response.WriteEntity(buildRun)
}

func (h *apiHandler) getImageBuildRun(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	buildrunName := request.PathParameter("imageBuildrun")

	// get imageBuildRun
	buildRun := buildv1alpha1.BuildRun{}
	if err := h.client.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: buildrunName}, &buildRun); err != nil {
		kapis.HandleError(request, response, err)
		return
	}
	_ = response.WriteEntity(&buildRun)
}

func (h *apiHandler) deleteImageBuildRun(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	buildrunName := request.PathParameter("imageBuildrun")
	ctx := context.Background()

	// get imageBuild
	buildRun := buildv1alpha1.BuildRun{}
	if err := h.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: buildrunName}, &buildRun); err != nil {
		kapis.HandleError(request, response, err)
		return
	}
	if err := h.client.Delete(context.Background(), &buildRun); err != nil {
		kapis.HandleError(request, response, err)
		return
	}
	_ = response.WriteEntity(&buildRun)
}

func (h *apiHandler) listImageBuildRuns(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	imageBuild := request.PathParameter("build")

	queryParam := query.ParseQueryParameter(request)
	labelSelector := labels.SelectorFromSet(labels.Set{"build.shipwright.io/name": imageBuild})

	opts := make([]client.ListOption, 0, 3)
	opts = append(opts, client.InNamespace(namespace))
	opts = append(opts, client.MatchingLabelsSelector{Selector: labelSelector})

	buildRunList := &buildv1alpha1.BuildRunList{}
	// fetch PipelineRuns
	if err := h.client.List(context.Background(), buildRunList, opts...); err != nil {
		kapis.HandleError(request, response, err)
		return
	}

	apiResult := resourcesV1alpha3.DefaultList(toBuildRunObjects(buildRunList.Items),
		queryParam,
		resourcesV1alpha3.DefaultCompare(),
		resourcesV1alpha3.DefaultFilter(), nil)

	_ = response.WriteAsJson(apiResult)
}

func toBuildRunObjects(apps []buildv1alpha1.BuildRun) []runtime.Object {
	objs := make([]runtime.Object, len(apps))
	for i := range apps {
		objs[i] = &apps[i]
	}
	return objs
}
