/*
Copyright 2024 The KubeSphere Authors.

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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/kubesphere/ks-devops/cmd/apiserver/app/options"
	"github.com/kubesphere/ks-devops/pkg/apiserver/runtime"
	swagger2 "github.com/kubesphere/ks-devops/pkg/apiserver/swagger"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

var output string

func init() {
	flag.StringVar(&output, "output", "./api/openapi-spec/swagger.json", "--output=./api.json")
}

func main() {
	flag.Parse()
	swaggerSpec := generateSwaggerJson()

	err := validateSpec(swaggerSpec)
	if err != nil {
		klog.Warningf("Swagger specification has errors")
	}
}

func validateSpec(apiSpec []byte) error {

	swaggerDoc, err := loads.Analyzed(apiSpec, "")
	if err != nil {
		return err
	}

	// Attempts to report about all errors
	validate.SetContinueOnErrors(true)

	v := validate.NewSpecValidator(swaggerDoc.Schema(), strfmt.Default)
	result, _ := v.Validate(swaggerDoc)

	if result.HasWarnings() {
		log.Printf("See warnings below:\n")
		for _, desc := range result.Warnings {
			log.Printf("- WARNING: %s\n", desc.Error())
		}

	}
	if result.HasErrors() {
		str := fmt.Sprintf("The swagger spec is invalid against swagger specification %s.\nSee errors below:\n", swaggerDoc.Version())
		for _, desc := range result.Errors {
			str += fmt.Sprintf("- %s\n", desc.Error())
		}
		log.Println(str)
		return errors.New(str)
	}

	return nil
}

func generateSwaggerJson() []byte {
	container := runtime.Container

	ctx := context.Background()
	serverRunOptions := options.NewServerRunOptions()
	server, err := serverRunOptions.NewAPIServer(ctx.Done())
	if err != nil {
		log.Fatal(err)
	}
	server.SetContainer(container)
	server.InstallDevOpsAPIs()

	config := swagger2.GetSwaggerConfig(container)

	swagger := restfulspec.BuildSwagger(config)

	data, _ := json.MarshalIndent(swagger, "", "  ")
	err = os.WriteFile(output, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully written to %s", output)

	return data
}
