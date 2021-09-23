package config

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
	"time"
)

// ControllerOptions is the option of Jenkins configuration controller
type ControllerOptions struct {
	LimitRangeClient    v1core.LimitRangesGetter
	ResourceQuotaClient v1core.ResourceQuotasGetter
	ConfigMapClient     v1core.ConfigMapsGetter

	ConfigMapInformer corev1informer.ConfigMapInformer
	NamespaceInformer corev1informer.NamespaceInformer

	InformerFactory informers.InformerFactory
	ConfigOperator  devops.ConfigurationOperator

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

	klog.Info("starting Jenkins config controller")
	defer klog.Info("shutting down Jenkins config controller")

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

// checkJenkinsConfigData makes sure that annotation devops.kubesphere.io/ks-jenkins-config exists
func (c *Controller) checkJenkinsConfigData(cm *v1.ConfigMap) (err error) {
	if data, ok := cm.Data[jenkinsYamlKey]; ok {
		if _, ok = cm.Data[jenkinsUserYamlKey]; !ok {
			cm.Data[jenkinsUserYamlKey] = data
		}
	}
	return
}

// checkJenkinsConfigFormula makes sure that annotation devops.kubesphere.io/jenkins-config-formula exists
func (c *Controller) checkJenkinsConfigFormula(cm *v1.ConfigMap) (err error) {
	annos := cm.GetAnnotations()
	if annos == nil {
		cm.Annotations = make(map[string]string)
		annos = cm.Annotations
	}

	if formulaName, ok := annos[ANNOJenkinsConfigFormula]; !ok || !isValidJenkinsConfigFormulaName(formulaName) {
		cm.Annotations[ANNOJenkinsConfigFormula] = FormulaCustom
		cm.Annotations[ANNOJenkinsConfigCustomized] = "true"
	}
	return
}

func isValidJenkinsConfigFormulaName(name string) bool {
	switch name {
	case FormulaCustom, FormulaHigh, FormulaLow:
		return true
	default:
		return false
	}
}

func (c *Controller) providePredefinedConfig(cm *v1.ConfigMap) (err error) {
	annos := cm.GetAnnotations()
	if annos == nil {
		return
	}

	if isJenkinsConfigCustomized(annos) {
		return
	}

	formula := annos[ANNOJenkinsConfigFormula]
	klog.Infof("Jenkins config formula name: %s", formula)

	var config map[string]string
	switch formula {
	case FormulaLow:
		config = getDefaultConfig()
	case FormulaHigh:
		config = getHighConfig()
	default:
		err = fmt.Errorf("invalid formula name: %s", formula)
		return
	}
	config[resourceLimitKey] = formula

	if err = c.handleWorkerNamespaceQuotaLimit(config, c.devopsOptions.WorkerNamespace); err != nil {
		err = fmt.Errorf("failed to handleWorkerNamespaceQuotaLimit, error: %v", err)
		return
	}
	if err = c.handleWorkerNamespaceLimitRange(config, c.devopsOptions.WorkerNamespace); err != nil {
		err = fmt.Errorf("failed to handleWorkerNamespaceLimitRange, error: %v", err)
		return
	}
	if err = c.handleJenkinsCasCConfig(cm, config); err != nil {
		err = fmt.Errorf("failed to handleJenkinsCasCConfig, error: %v", err)
		return
	}
	return
}

func isJenkinsConfigCustomized(annos map[string]string) bool {
	if val, ok := annos[ANNOJenkinsConfigFormula]; ok && val == FormulaCustom {
		return true
	}
	if val, ok := annos[ANNOJenkinsConfigCustomized]; ok && val == "true" {
		return true
	}
	return false
}

func (c *Controller) reloadJenkinsConfig() (err error) {
	if c.configOperator == nil {
		err = fmt.Errorf("failed to reload Jenkins config due to the configOperator is nil")
	} else {
		err = c.configOperator.ApplyNewSource(fmt.Sprintf("/var/jenkins_home/casc_configs/%s", jenkinsUserYamlKey))
	}
	return
}

func (c *Controller) delayReloadJenkinsConfig() (err error) {
	// Wait some period to reload
	if c.ReloadCasCDelay > 0 {
		klog.V(5).Infof("waiting %s to reload Jenkins configuration", c.ReloadCasCDelay.String())
		time.Sleep(c.ReloadCasCDelay)
	}
	return c.reloadJenkinsConfig()
}

func (c *Controller) getConfigMapFromKey(key string) (cm *v1.ConfigMap, err error) {
	var (
		ns   string
		name string
	)
	if ns, name, err = cache.SplitMetaNamespaceKey(key); err != nil {
		err = fmt.Errorf("could not split meta namespace key %s", key)
		return
	}

	_, err = c.namespaceLister.Get(ns)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(6).Infof("namespace %s in work queue no longer exists", ns)
			err = nil
			return
		}
		err = fmt.Errorf("could not get namespace: %s, error is: %v", ns, err)
		return
	}

	cm, err = c.configmapLister.ConfigMaps(ns).Get(name)
	return
}

func (c *Controller) syncHandler(key string) (err error) {
	klog.Info("syncing key:", key)
	defer klog.Info("synced key: ", key)
	var jenkinsCM *v1.ConfigMap
	if jenkinsCM, err = c.getConfigMapFromKey(key); err != nil {
		return
	}

	jenkinsCMCopy := jenkinsCM.DeepCopy()
	ns, name := jenkinsCMCopy.Namespace, jenkinsCMCopy.Name

	if err = c.checkJenkinsConfigData(jenkinsCMCopy); err != nil {
		err = fmt.Errorf("failed with checking Jenkins ConfigMap data, error: %v", err)
		return
	}

	if err = c.checkJenkinsConfigFormula(jenkinsCMCopy); err != nil {
		err = fmt.Errorf("failed with checking Jenkins config forumla, error: %v", err)
		return
	}

	if err = c.providePredefinedConfig(jenkinsCMCopy); err != nil {
		err = fmt.Errorf("failed to provide the pre-defined Jenkins config, error: %v", err)
		return
	}

	// Update Jenkins Configuration as Code ConfigMap
	_, err = c.configMapClient.ConfigMaps(ns).Update(context.Background(), jenkinsCMCopy, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("failed to update ConfigMap: %s/%s", ns, name)
		return err
	}

	// Reload configuration
	klog.V(5).Info("reloading Jenkins configuration")
	if err = c.delayReloadJenkinsConfig(); err == nil {
		klog.V(5).Infof("reloaded Jenkins configuration successfully")
	}
	return
}

// Handle worker namespace quota limit
func (c *Controller) handleWorkerNamespaceQuotaLimit(providedConfig map[string]string, namespace string) error {
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
func (c *Controller) handleWorkerNamespaceLimitRange(providedConfig map[string]string, namespace string) error {
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
func (c *Controller) handleJenkinsCasCConfig(cm *v1.ConfigMap, providedConfig map[string]string) (err error) {
	jenkinsCasCConfigTemplate := cm.Data[jenkinsYamlKey]
	namespace := cm.Namespace

	cascMap := make(map[string]interface{})
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

	// set CasC config into newJenkinsCascConfig
	var targetJenkinsYAMLConfig []byte
	if targetJenkinsYAMLConfig, err = yaml.Marshal(cascMap); err == nil {
		// update Jenkins config ConfigMap
		cm.Data[jenkinsUserYamlKey] = string(targetJenkinsYAMLConfig)
	} else {
		return
	}

	if _, err = c.configMapClient.ConfigMaps(namespace).Update(context.Background(), cm, metav1.UpdateOptions{}); err != nil {
		err = fmt.Errorf("failed to update ConfigMap: %s/%s, error: %v", namespace, jenkinsCasCConfigName, err)
	} else {
		klog.V(5).Infof("updated ConfigMap %s/%s successfully", namespace, jenkinsCasCConfigName)
	}
	return
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
