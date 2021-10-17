package pipeline

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/jenkins-zh/jenkins-client/pkg/core"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	// MetaUpdated indicates the metadata of Pipeline has updated into annotations of Pipeline CR.
	MetaUpdated = "MetaUpdated"
	// FailedMetaUpdate indicates the controller fails to update metadata of Pipeline.
	FailedMetaUpdate = "FailedMetaUpdate"
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
	log := r.log.WithValues("Pipeline", req.NamespacedName)
	pipeline := &v1alpha3.Pipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
		// ignore resource not found due to deletion
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.obtainAndUpdatePipelineMetadata(pipeline); err != nil {
		log.Error(err, "unable to obtain and update Pipeline metadata from Jenkins")
		r.onFailedMetaUpdate(pipeline, err)
		return ctrl.Result{}, err
	}

	if err := r.obtainAndUpdatePipelineBranches(pipeline); err != nil {
		log.Error(err, "unable to obtain and update Pipeline branches data from Jenkins")
		r.onFailedMetaUpdate(pipeline, err)
		return ctrl.Result{}, err
	}

	if err := r.updateAnnotations(pipeline.Annotations, req.NamespacedName); err != nil {
		log.Error(err, "unable to update annotations of Pipeline")
		return ctrl.Result{}, err
	}

	// re-synch after 10 seconds
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *Reconciler) onFailedMetaUpdate(pipeline *v1alpha3.Pipeline, err error) {
	r.recorder.Eventf(pipeline, v1.EventTypeWarning, FailedMetaUpdate, "Failed to update metadata of Pipeline from Jenkins, err = %v", err)
}

func (r *Reconciler) onUpdateMetaSuccessfully(pipeline *v1alpha3.Pipeline) {
	r.recorder.Eventf(pipeline, v1.EventTypeNormal, MetaUpdated, "Metadata of Pipeline has been updated from Jenkins successfully")
}

func (r *Reconciler) obtainAndUpdatePipelineMetadata(pipeline *v1alpha3.Pipeline) error {
	boClient := job.BlueOceanClient{
		JenkinsCore:  r.JenkinsCore,
		Organization: "jenkins",
	}
	// fetch pipeline metadata from Jenkins
	jobPipeline, err := boClient.GetPipeline(pipeline.Name, pipeline.Namespace)
	if err != nil {
		return err
	}
	metadataJSON, err := json.Marshal(convertPipeline(jobPipeline))
	if err != nil {
		return err
	}
	// update annotation
	pipeline.Annotations[v1alpha3.PipelineJenkinsMetadataAnnoKey] = string(metadataJSON)
	return nil
}

func (r *Reconciler) obtainAndUpdatePipelineBranches(pipeline *v1alpha3.Pipeline) error {
	if pipeline.Spec.Type != v1alpha3.MultiBranchPipelineType {
		// skip non multi-branch Pipeline
		return nil
	}
	boClient := job.BlueOceanClient{
		JenkinsCore:  r.JenkinsCore,
		Organization: "jenkins",
	}
	jobBranches, err := boClient.GetBranches(job.GetBranchesOption{
		Folders:      []string{pipeline.Namespace},
		PipelineName: pipeline.Name,
		Limit:        100000,
	})
	if err != nil {
		return err
	}

	branchesJSON, err := json.Marshal(convertBranches(jobBranches))
	if err != nil {
		return err
	}

	// update annotation
	pipeline.Annotations[v1alpha3.PipelineJenkinsBranchesAnnoKey] = string(branchesJSON)
	return nil
}

func (r *Reconciler) updateAnnotations(annotations map[string]string, pipelineKey client.ObjectKey) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pipeline := &v1alpha3.Pipeline{}
		if err := r.Get(context.Background(), pipelineKey, pipeline); err != nil {
			return client.IgnoreNotFound(err)
		}

		if reflect.DeepEqual(pipeline.Annotations, annotations) {
			return nil
		}

		// update annotations
		pipeline.Annotations = annotations

		err := r.Update(context.Background(), pipeline)
		if err == nil {
			r.onUpdateMetaSuccessfully(pipeline)
		}
		return err
	})
}

// pipelineMetadataPredicate returns a predicate which only cares about CreateEvent.
var pipelineMetadataPredicate = predicate.Funcs{
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
		return false
	},
}

// SetupWithManager setups reconciler with controller manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("pipeline-metadata-controller")
	r.log = ctrl.Log.WithName("pipeline-metadata-controller")
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(pipelineMetadataPredicate).
		For(&v1alpha3.Pipeline{}).
		Complete(r)
}
