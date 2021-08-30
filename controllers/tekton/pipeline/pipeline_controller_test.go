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

package pipeline

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	devopsv2alpha1 "kubesphere.io/devops/pkg/api/devops/v2alpha1"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("Pipeline controller", func() {

	// Here we define utility constants for object names and testing timeouts/durations and intervals.
	const (
		DevopsPipelineName      = "test-pipeline"
		DevopsPipelineNamespace = "default"
		DevopsTaskName          = "test-task"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating CronJob Status", func() {
		It("Should create Devops Pipeline and Tekton Pipeline", func() {
			By("By creating a new Devops Pipeline")
			ctx := context.Background()
			devopsPipeline := &devopsv2alpha1.Pipeline{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "devops.kubesphere.io/v2alpha1",
					Kind:       "Pipeline",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      DevopsPipelineName,
					Namespace: DevopsPipelineNamespace,
				},
				Spec: devopsv2alpha1.PipelineSpec{
					Tasks: []devopsv2alpha1.TaskSpec{
						{
							Name: "test-task-1",
							Steps: []devopsv2alpha1.Step{
								{
									Name:    "test-step",
									Image:   "ubuntu:latest",
									Command: []string{"echo"},
									Args:    []string{"step", "test"},
								},
							},
						},
						{
							Name: "test-task-2",
							Steps: []devopsv2alpha1.Step{
								{
									Name:   "test-step",
									Image:  "alpine:latest",
									Script: "date\nls",
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, devopsPipeline)).Should(Succeed())

			devopPipelineLookupKey := types.NamespacedName{Name: DevopsPipelineName, Namespace: DevopsPipelineNamespace}
			createdDevopsPipeline := &devopsv2alpha1.Pipeline{}

			// We'll need to retry getting this newly created Devops Pipeline, given that creation may not immediately happen.
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, devopPipelineLookupKey, createdDevopsPipeline); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			// We then need to verify the creation of Tekton Pipeline and Tasks.
			// This block is to verify creation of Tekton Tasks.
			Eventually(func() bool {
				for _, devopsTask := range devopsPipeline.Spec.Tasks {
					tknTaskLookupKey := types.NamespacedName{Name: DevopsPipelineName + "-" + devopsTask.Name, Namespace: DevopsPipelineNamespace}
					createdTknTask := &tektonv1.Task{}
					if err := k8sClient.Get(ctx, tknTaskLookupKey, createdTknTask); err != nil {
						return false
					}
				}
				return true
			}, timeout, interval).Should(BeTrue())

			// This block is to verify creation of Tekton Pipeline.
			Eventually(func() bool {
				tknPipelineLookupKey := types.NamespacedName{Name: DevopsPipelineName, Namespace: DevopsPipelineNamespace}
				createdTknPipeline := &tektonv1.Pipeline{}
				if err := k8sClient.Get(ctx, tknPipelineLookupKey, createdTknPipeline); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})
	})

})

/*
	After writing all this code, you can run `go test ./...` in your `controllers/` directory again to run your new test!
*/
