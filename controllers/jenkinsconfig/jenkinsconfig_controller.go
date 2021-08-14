package jenkinsconfig

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1informer "k8s.io/client-go/informers/core/v1"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"kubesphere.io/devops/pkg/client/devops"
	"kubesphere.io/devops/pkg/client/devops/jenkins"
	"kubesphere.io/devops/pkg/informers"
	"strconv"
	"strings"
	"time"
)

// ControllerOptions devops config controller options
type ControllerOptions struct {
	LimitRangeClient    v1core.LimitRangesGetter
	ResourceQuotaClient v1core.ResourceQuotasGetter
	ConfigMapClient     v1core.ConfigMapsGetter

	ConfigMapInformer corev1informer.ConfigMapInformer
	NamespaceInformer corev1informer.NamespaceInformer

	InformerFactory informers.InformerFactory

	ConfigOperator devops.ConfigurationOperator

	ReloadCasCDelay time.Duration
}

// Controller is used to maintain the state of the jenkins-casc-config ConfigMap.
type Controller struct {
	configmapLister corev1lister.ConfigMapLister
	namespaceLister corev1lister.NamespaceLister
	configmapSynced cache.InformerSynced

	limitRangeClient    v1core.LimitRangesGetter
	resourceQuotaClient v1core.ResourceQuotasGetter
	configMapClient     v1core.ConfigMapsGetter

	configOperator devops.ConfigurationOperator

	queue            workqueue.RateLimitingInterface
	workerLoopPeriod time.Duration
	ReloadCasCDelay  time.Duration

	devopsOptions *jenkins.Options
}

// NewController creates a new JenkinsConfigController
func NewController(
	options *ControllerOptions,
	devopsOptions *jenkins.Options,
) *Controller {
	controller := &Controller{
		configmapLister: options.ConfigMapInformer.Lister(),
		namespaceLister: options.NamespaceInformer.Lister(),
		configmapSynced: options.ConfigMapInformer.Informer().HasSynced,

		limitRangeClient:    options.LimitRangeClient,
		resourceQuotaClient: options.ResourceQuotaClient,
		configMapClient:     options.ConfigMapClient,

		configOperator: options.ConfigOperator,

		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), jenkinsConfigName),
		workerLoopPeriod: time.Second,
		ReloadCasCDelay:  options.ReloadCasCDelay,

		devopsOptions: devopsOptions,
	}

	options.ConfigMapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueue,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldConfigMap := oldObj.(*v1.ConfigMap)
			newConfigMap := newObj.(*v1.ConfigMap)
			if oldConfigMap.ResourceVersion == newConfigMap.ResourceVersion {
				return
			}
			controller.enqueue(newObj)
		},
		DeleteFunc: controller.enqueue,
	})
	return controller
}

func (c *Controller) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	// Filter by namespace and name
	if namespace, name, _ := cache.SplitMetaNamespaceKey(key); namespace != c.devopsOptions.Namespace || name != jenkinsConfigName {
		return
	}
	c.queue.Add(key)
}

func (c *Controller) worker() {
	for c.processNextWorkItem() {
	}
}

// Start implements for Runnable interface.
func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.run(1, stopCh)
}

func (c *Controller) run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("starting jenkinsconfig controller")
	defer klog.Info("shutting down jenkinsconfig controller")

	if !cache.WaitForCacheSync(stopCh, c.configmapSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, c.workerLoopPeriod, stopCh)
	}

	<-stopCh
	return nil
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.queue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.queue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in queue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.queue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.queue.Forget(obj)
		klog.V(5).Infof("successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	klog.Info("syncing key:", key)
	defer klog.Info("synced key: ", key)
	namespace, configMapName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Error(fmt.Sprintf("could not split meta namespace key %s", key))
		return nil
	}

	_, err = c.namespaceLister.Get(namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Error(fmt.Sprintf("namespace %s in work queue no longer exists", namespace))
			return nil
		}
		klog.Error(fmt.Sprintf("could not get namespace: %s", namespace))
		return err
	}

	devopsConfig, err := c.configmapLister.ConfigMaps(namespace).Get(configMapName)
	if err != nil {
		return err
	}

	podResourceLimit := devopsConfig.DeepCopy().Data[podResourceLimitConfigName]

	klog.Infof("Jenkins agent pod resource limit: %q", podResourceLimit)

	// check if the customized label is labeled
	customizedLabeledRaw, ok := devopsConfig.ObjectMeta.GetLabels()[customizedLabel]
	var isCustomizedLabeled bool
	if ok {
		isCustomizedLabeled, err = strconv.ParseBool(customizedLabeledRaw)
		if err != nil {
			klog.Warningf("ConfigMap: %q has invalid label: %q, error: %q", key, customizedLabel, err)
			// And do nothing
			return nil
		}
	}

	var config map[string]string
	if strings.Compare(string(customLimit), podResourceLimit) == 0 || isCustomizedLabeled {
		// get configuration
		config = devopsConfig.Data

		// label the configmap
		devopsConfig.SetLabels(map[string]string{
			customizedLabel: "true",
		})
		config[resourceLimitKey] = string(customLimit)
	} else if strings.Compare(string(highLimit), podResourceLimit) == 0 {
		// get configuration
		config = getHighConfig()
		config[resourceLimitKey] = string(highLimit)
	} else if isDefaultResourceLimit(podResourceLimit) {
		// get configuration
		config = getDefaultConfig()
		config[resourceLimitKey] = string(defaultLimit)
	}

	if len(config) > 0 {
		err := handleWorkerNamespaceQuotaLimit(c, config, c.devopsOptions.WorkerNamespace)
		if err != nil {
			return err
		}
		err = handleWorkerNamespaceLimitRange(c, config, c.devopsOptions.WorkerNamespace)
		if err != nil {
			return err
		}
		err = handleJenkinsCasCConfig(c, config, namespace)
		if err != nil {
			return err
		}
	}

	// Update jenkinsconfig ConfigMap
	_, err = c.configMapClient.ConfigMaps(namespace).Update(context.Background(), devopsConfig, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("failed to update ConfigMap: %s/%s", namespace, configMapName)
		return err
	}

	// Wait some period to reload
	klog.V(5).Infof("waiting %s to reload Jenkins configuration", c.ReloadCasCDelay.String())
	time.Sleep(c.ReloadCasCDelay)

	// Reload configuration
	klog.V(5).Info("reloading Jenkins configuration")
	err = c.configOperator.ReloadConfiguration()
	if err != nil {
		return err
	}
	klog.V(5).Infof("reloaded Jenkins configuration successfully")

	return nil
}

// Handle worker namespace quota limit
func handleWorkerNamespaceQuotaLimit(c *Controller, providedConfig map[string]string, namespace string) error {
	// get the resource quota
	workerResourceQuota, err := c.resourceQuotaClient.ResourceQuotas(namespace).Get(context.Background(), workerResQuotaName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("could not find ResourceQuota: %s/%s", namespace, workerResQuotaName)
		return err
	}
	// apply new changes
	newWorkerResourceQuota := workerResourceQuota.DeepCopy()
	limitCPURaw, ok := providedConfig[workerLimitCPUKey]
	if ok {
		limitCPU, err := resource.ParseQuantity(limitCPURaw)
		if err != nil {
			// invalid limit cpu
			klog.Errorf("failed to parse limit.cpu: %s", limitCPURaw)
			return err
		}
		newWorkerResourceQuota.Spec.Hard["limits.cpu"] = limitCPU
	}

	limitMemoryRaw, ok := providedConfig[workerLimitMemoryKey]
	if ok {
		limitMemory, err := resource.ParseQuantity(limitMemoryRaw)
		if err != nil {
			klog.Errorf("failed to parse limit.memory: %s", limitMemoryRaw)
			return err
		}
		newWorkerResourceQuota.Spec.Hard["limits.memory"] = limitMemory
	}

	// update worker resource quota
	_, err = c.resourceQuotaClient.ResourceQuotas(namespace).Update(context.Background(), newWorkerResourceQuota, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("failed to update ResourceQuota: %s/%s", namespace, workerResQuotaName)
		return err
	}
	klog.V(5).Infof("updated %s/%s successfully", namespace, workerResQuotaName)
	return nil
}

// Handle worker namespace limit range
func handleWorkerNamespaceLimitRange(c *Controller, providedConfig map[string]string, namespace string) error {
	workerLimitRange, err := c.limitRangeClient.LimitRanges(namespace).Get(context.Background(), workerLimitRangeName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("could not find LimitRange: %s/%s", namespace, workerLimitRangeName)
		return err
	}
	newWorkerLimitRange := workerLimitRange.DeepCopy()
	for _, limit := range newWorkerLimitRange.Spec.Limits {
		// handle for container type
		if v1.LimitTypeContainer == limit.Type {
			// handle default cpu
			if defaultCPURaw, ok := providedConfig[workerLRDefaultCPUKey]; ok {
				defaultCPU, err := resource.ParseQuantity(defaultCPURaw)
				if err != nil {
					klog.Errorf("failed to parse default.cpu: %s", defaultCPURaw)
					return err
				}
				limit.Default["cpu"] = defaultCPU
			}

			// handle default memory
			if defaultMemoryRaw, ok := providedConfig[workerLRDefaultMemoryKey]; ok {
				defaultMemory, err := resource.ParseQuantity(defaultMemoryRaw)
				if err != nil {
					klog.Errorf("failed to parse default.memory: %s", defaultMemoryRaw)
					return err
				}
				limit.Default["memory"] = defaultMemory
			}

			// handle default request cpu
			if defaultRequestCPURaw, ok := providedConfig[workerLRDefaultReqCPUKey]; ok {
				defaultRequestCPU, err := resource.ParseQuantity(defaultRequestCPURaw)
				if err != nil {
					klog.Errorf("failed to parse defaultrequest.cpu: %s", defaultRequestCPURaw)
					return err
				}
				limit.DefaultRequest["cpu"] = defaultRequestCPU
			}

			// handle default request memory
			if defaultRequestMemoryRaw, ok := providedConfig[workerLRDefaultReqMemoryKey]; ok {
				defaultRequestMemory, err := resource.ParseQuantity(defaultRequestMemoryRaw)
				if err != nil {
					klog.Errorf("failed to parse defaultrequest.memory: %s", defaultRequestMemoryRaw)
					return err
				}
				limit.DefaultRequest["memory"] = defaultRequestMemory
			}
		}
	}

	// update worker limit range
	_, err = c.limitRangeClient.LimitRanges(namespace).Update(context.Background(), newWorkerLimitRange, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("failed to update %s/%s", namespace, workerLimitRangeName)
		return err
	}
	klog.V(5).Infof("updated %s/%s successfully", namespace, workerLimitRangeName)
	return nil
}

// Handle Jenkins Configuration-as-Code configuration
func handleJenkinsCasCConfig(c *Controller, providedConfig map[string]string, namespace string) error {
	// get CasC configuration
	jenkinsCasCConfig, err := c.configmapLister.ConfigMaps(namespace).Get(jenkinsCasCConfigName)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("could not find %s/%s", namespace, jenkinsCasCConfigName)
		return err
	}
	jenkinsCascConfigToUpdate := jenkinsCasCConfig.DeepCopy()

	cascMap := make(map[string]interface{})

	jenkinsCasCConfigTemplate, err := getOrCreateJenkinsCasCTemplate(c.configMapClient, namespace, jenkinsCascConfigToUpdate.Data[jenkinsYamlKey])
	if err != nil {
		return err
	}
	// overwrite Jenkins CasC config template only while custom resource limit set
	if strings.Compare(providedConfig[resourceLimitKey], string(customLimit)) == 0 {
		if customJenkinsYaml, ok := providedConfig[jenkinsYamlKey]; ok {
			jenkinsCasCConfigTemplate = customJenkinsYaml
		}
	}
	err = yaml.Unmarshal([]byte(jenkinsCasCConfigTemplate), &cascMap)
	if err != nil {
		return err
	}

	kubernetesMap := cascMap["jenkins"].(map[interface{}]interface{})["clouds"].([]interface{})[0].(map[interface{}]interface{})["kubernetes"]

	// set concurrent
	concurrent, ok := providedConfig[podConcurrentKey]
	if ok {
		kubernetesMap.(map[interface{}]interface{})["containerCapStr"] = concurrent
	}

	// set pod template
	if templates, ok := kubernetesMap.(map[interface{}]interface{})["templates"].([]interface{}); ok {
		for _, template := range templates {
			// type safe check
			if template, ok := template.(map[interface{}]interface{}); ok {
				// ensure template name exist
				if name, ok := template["name"]; ok {
					if containers, ok := template["containers"]; ok {
						switch name {
						case "base":
							// handle base pod template
							setContainersLimit(providedConfig, containers, "base")
						case "nodejs":
							// handle nodejs pod template
							setContainersLimit(providedConfig, containers, "nodejs")
						case "maven":
							// handle maven pod template
							setContainersLimit(providedConfig, containers, "maven")
						case "go":
							// handle go pod template
							setContainersLimit(providedConfig, containers, "go")
						}
					}
				}
			}
		}
	}

	// set casc config into newJenkinsCascConfig
	jenkinsYaml, err := yaml.Marshal(cascMap)
	if err != nil {
		return err
	}

	// Update jenkinsconfig configmap
	jenkinsCascConfigToUpdate.Data[jenkinsYamlKey] = string(jenkinsYaml)
	_, err = c.configMapClient.ConfigMaps(namespace).Update(context.Background(), jenkinsCascConfigToUpdate, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("failed to update configmap: %s/%s", namespace, jenkinsCasCConfigName)
		return err
	}

	klog.V(5).Infof("updated configmap %s/%s successfully", namespace, jenkinsCasCConfigName)
	return nil
}

// Set containers limit
func setContainersLimit(providedConfig map[string]string, containers interface{}, containerName string) {
	if containers, ok := containers.([]interface{}); ok {
		for _, container := range containers {
			if container, ok := container.(map[interface{}]interface{}); ok {
				if name, ok := container["name"]; ok {
					switch name {
					case containerName:
						// set specified container resource limit
						if limitCPU, ok := providedConfig[containerName+".limit.cpu"]; ok {
							container["resourceLimitCpu"] = limitCPU
						}
						if limitMemory, ok := providedConfig[containerName+".limit.memory"]; ok {
							container["resourceLimitMemory"] = limitMemory
						}
					case "jnlp":
						// set jnlp container resource limit
						if limitCPU, ok := providedConfig[jnlpLimitCPUKey]; ok {
							container["resourceLimitCpu"] = limitCPU
						}
						if limitMemory, ok := providedConfig[jnlpLimitMemoryKey]; ok {
							container["resourceLimitMemory"] = limitMemory
						}
					}
				}
			}
		}
	}
}
