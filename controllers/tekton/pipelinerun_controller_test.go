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

package tekton

// Potential integration test template
// import (
// 	"context"
// 	"time"

// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/types"

// 	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
// 	devopsv2alpha1 "kubesphere.io/devops/pkg/api/devops/v2alpha1"
// )

// // +kubebuilder:docs-gen:collapse=Imports

// var _ = Describe("PipelineRun controller", func() {

// 	// Here we define utility constants for object names and testing timeouts/durations and intervals.
// 	const (
// 		DevopsPipelineRunName      = "test-pipelinerun"
// 		DevopsPipelineRefName      = "test-pipeline"
// 		DevopsPipelineRunNamespace = "default"

// 		timeout  = time.Second * 10
// 		interval = time.Millisecond * 250
// 	)

// 	Context("When updating PipelineRun Status", func() {
// 		It("Should create Devops PipelineRun", func() {
// 			By("By creating a new Devops PipelineRun")
// 			ctx := context.Background()
// 			devopsPipelineRun := &devopsv2alpha1.PipelineRun{
// 				TypeMeta: metav1.TypeMeta{
// 					APIVersion: "devops.kubesphere.io/v2alpha1",
// 					Kind:       "PipelineRun",
// 				},
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      DevopsPipelineRunName,
// 					Namespace: DevopsPipelineRunNamespace,
// 				},
// 				Spec: devopsv2alpha1.PipelineRunSpec{
// 					Name:        DevopsPipelineRunName,
// 					PipelineRef: DevopsPipelineRefName,
// 				},
// 			}
// 			Expect(k8sClient.Create(ctx, devopsPipelineRun)).Should(Succeed())

// 			devopPipelineRunLookupKey := types.NamespacedName{Name: DevopsPipelineRunName, Namespace: DevopsPipelineRunNamespace}
// 			createdDevopsPipelineRun := &devopsv2alpha1.PipelineRun{}

// 			// We'll need to retry getting this newly created Devops PipelineRun, given that creation may not immediately happen.
// 			Eventually(func() bool {
// 				if err := k8sClient.Get(ctx, devopPipelineRunLookupKey, createdDevopsPipelineRun); err != nil {
// 					return false
// 				}
// 				return true
// 			}, timeout, interval).Should(BeTrue())

// 			// We then need to verify the creation of Tekton PipelineRun.
// 			Eventually(func() bool {
// 				tknPipelineRunLookupKey := types.NamespacedName{Name: DevopsPipelineRunName, Namespace: DevopsPipelineRunNamespace}
// 				tknPipelineRun := &tektonv1.PipelineRun{}
// 				if err := k8sClient.Get(ctx, tknPipelineRunLookupKey, tknPipelineRun); err != nil {
// 					return false
// 				}
// 				return true
// 			}, timeout, interval).Should(BeTrue())
// 		})
// 	})

// })
