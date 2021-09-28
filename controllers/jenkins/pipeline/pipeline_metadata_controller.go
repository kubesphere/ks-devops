package pipeline

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-logr/logr"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Reconciler reconciles metadata of Pipeline.
type Reconciler struct {
	client.Client
	JenkinsCore core.JenkinsCore
	recorder    record.EventRecorder
	log         logr.Logger
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=pipelines/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	// log := r.log.WithValues("Pipeline", req.NamespacedName)
	pipeline := &v1alpha3.Pipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
		// ignore resource not found due to deletion
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// fetch pipeline metadata from Jenkins
	boClient := job.BlueOceanClient{
		JenkinsCore:  r.JenkinsCore,
		Organization: "jenkins",
	}
	pipelineMetadata, err := boClient.GetPipeline(pipeline.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update pipeline metadata
	if err := r.updateMetadata(pipelineMetadata, req.NamespacedName); err != nil {
		return ctrl.Result{}, err
	}

	// re-synch after 10 seconds
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *Reconciler) updateMetadata(metadata *job.Pipeline, pipelineKey client.ObjectKey) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pipeline := &v1alpha3.Pipeline{}
		if err := r.Get(context.Background(), pipelineKey, pipeline); err != nil {
			return client.IgnoreNotFound(err)
		}
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		// diff pipeline metadata
		if pipeline.Annotations[v1alpha3.PipelineJenkinsMetadataAnnoKey] == string(metadataJSON) {
			// skip updation if metadata unchanged
			return nil
		}
		// update annotations
		if pipeline.Annotations == nil {
			pipeline.Annotations = map[string]string{}
		}
		pipeline.Annotations[v1alpha3.PipelineJenkinsMetadataAnnoKey] = string(metadataJSON)
		return r.Update(context.Background(), pipeline)
	})
}

// getPipelineMetadataPredicate returns a predicate which only cares CreateEvent and GenericEvent.
func getPipelineMetadataPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(ce event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(de event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(ue event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(ge event.GenericEvent) bool {
			return true
		},
	}
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("pipeline-metadata-controller")
	r.log = ctrl.Log.WithName("pipeline-metadata-controller")
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(getPipelineMetadataPredicate()).
		For(&v1alpha3.Pipeline{}).
		Complete(r)
}
