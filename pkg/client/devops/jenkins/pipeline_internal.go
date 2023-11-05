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

package jenkins

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"kubesphere.io/devops/pkg/client/devops/jenkins/triggers"

	"github.com/beevik/etree"

	devopsv1alpha3 "kubesphere.io/devops/pkg/api/devops/v1alpha3"

	"kubesphere.io/devops/pkg/client/devops/jenkins/internal"
)

func replaceXmlVersion(config, oldVersion, targetVersion string) string {
	lines := strings.Split(config, "\n")
	lines[0] = strings.Replace(lines[0], oldVersion, targetVersion, -1)
	output := strings.Join(lines, "\n")
	return output
}

func createPipelineConfigXml(pipeline *devopsv1alpha3.NoScmPipeline) (string, error) {
	doc := etree.NewDocument()
	xmlString := `<?xml version='1.0' encoding='UTF-8'?>
<flow-definition plugin="workflow-job">
  <actions>
    <org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobAction plugin="pipeline-model-definition"/>
    <org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobPropertyTrackerAction plugin="pipeline-model-definition">
      <jobProperties/>
      <triggers/>
      <parameters/>
      <options/>
    </org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobPropertyTrackerAction>
  </actions>
</flow-definition>
`
	doc.ReadFromString(xmlString)
	flow := doc.SelectElement("flow-definition")
	flow.CreateElement("description").SetText(pipeline.Description)
	properties := flow.CreateElement("properties")

	if pipeline.DisableConcurrent {
		properties.CreateElement("org.jenkinsci.plugins.workflow.job.properties.DisableConcurrentBuildsJobProperty")
	}

	if pipeline.Discarder != nil {
		discarder := properties.CreateElement("jenkins.model.BuildDiscarderProperty")
		strategy := discarder.CreateElement("strategy")
		strategy.CreateAttr("class", "hudson.tasks.LogRotator")
		strategy.CreateElement("daysToKeep").SetText(pipeline.Discarder.DaysToKeep)
		strategy.CreateElement("numToKeep").SetText(pipeline.Discarder.NumToKeep)
		strategy.CreateElement("artifactDaysToKeep").SetText("-1")
		strategy.CreateElement("artifactNumToKeep").SetText("-1")
	}
	if pipeline.Parameters != nil {
		replaceParametersInEtree(properties, pipeline.Parameters)
	}

	// create trigger xml structure
	if pipeline.TimerTrigger != nil || pipeline.GenericWebhook != nil {
		triggersEle := properties.
			CreateElement("org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty").
			CreateElement("triggers")

		if pipeline.TimerTrigger != nil {
			triggersEle.CreateElement("hudson.triggers.TimerTrigger").CreateElement("spec").
				SetText(pipeline.TimerTrigger.Cron)
		}

		triggers.CreateGenericWebhookXML(triggersEle, pipeline.GenericWebhook)
	}

	pipelineDefine := flow.CreateElement("definition")
	pipelineDefine.CreateAttr("class", "org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition")
	pipelineDefine.CreateAttr("plugin", "workflow-cps")
	pipelineDefine.CreateElement("script").SetText(pipeline.Jenkinsfile)

	pipelineDefine.CreateElement("sandbox").SetText("true")

	flow.CreateElement("triggers")

	if pipeline.RemoteTrigger != nil {
		flow.CreateElement("authToken").SetText(pipeline.RemoteTrigger.Token)
	}
	flow.CreateElement("disabled").SetText("false")

	doc.Indent(2)
	stringXml, err := doc.WriteToString()
	if err != nil {
		return "", err
	}
	return replaceXmlVersion(stringXml, "1.0", "1.1"), err
}

func updatePipelineConfigXml(config string, pipeline *devopsv1alpha3.NoScmPipeline) (string, error) {
	config = replaceXmlVersion(config, "1.1", "1.0")
	doc := etree.NewDocument()
	err := doc.ReadFromString(config)
	if err != nil {
		return "", err
	}
	flow := doc.SelectElement(FlowTag)
	// ------------------------------------------------
	// update properties
	properties := flow.SelectElement(PropertiesTag)
	if pipeline.DisableConcurrent {
		addOrUpdateElement(properties, DisableConcurrentJobTag, StringNull)
	} else {
		removeChildElement(properties, DisableConcurrentJobTag)
	}

	if pipeline.Discarder != nil {
		var discarder, strategy *etree.Element
		discarder = addOrUpdateElement(properties, BuildDiscarderTag, StringNull)
		strategy = addOrUpdateElement(discarder, StrategyTag, StringNull)

		replaceAttr(strategy, ClassKey, "hudson.tasks.LogRotator")
		addOrUpdateElement(strategy, DaysToKeepTag, pipeline.Discarder.DaysToKeep)
		addOrUpdateElement(strategy, NumToKeepTag, pipeline.Discarder.NumToKeep)
		addOrUpdateElement(strategy, ArtiDaysToKeepTag, "-1")
		addOrUpdateElement(strategy, ArtiNumToKeepTag, "-1")
	} else {
		removeChildElement(properties, BuildDiscarderTag)
	}

	if pipeline.Parameters != nil { // overwrite parameters
		replaceParametersInEtree(properties, pipeline.Parameters)
	} else {
		removeChildElement(properties, ParamDefiPropTag)
	}

	// update triggers xml structure
	var pipelineTriggerEle, triggersEle *etree.Element
	pipelineTriggerEle = addOrUpdateElement(properties, PipelineTriggersJobTag, StringNull)
	triggersEle = addOrUpdateElement(pipelineTriggerEle, TriggersTag, StringNull)

	if pipeline.TimerTrigger != nil {
		timerTriggerEle := addOrUpdateElement(triggersEle, TimerTriggerTag, StringNull)
		addOrUpdateElement(timerTriggerEle, "spec", pipeline.TimerTrigger.Cron)
	} else {
		removeChildElement(triggersEle, TimerTriggerTag)
	}

	if pipeline.GenericWebhook != nil {
		// TODO issue: if support GenericWebhook in console, need to delete GenericWebhook tag when pipeline.GenericWebhook is nil;
		triggers.CreateGenericWebhookXML(triggersEle, pipeline.GenericWebhook)
	}

	// ------------------------------------------------
	// replace definition(all fields could update from console)
	removeChildElement(flow, DefinitionTag)
	pipelineDefine := flow.CreateElement(DefinitionTag)
	pipelineDefine.CreateAttr(ClassKey, "org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition")
	pipelineDefine.CreateAttr(PluginKey, "workflow-cps")
	pipelineDefine.CreateElement(ScriptTag).SetText(pipeline.Jenkinsfile)
	pipelineDefine.CreateElement(SandboxTag).SetText("true")

	// ------------------------------------------------
	// update others
	if flow.SelectElement(TriggersTag) == nil {
		flow.CreateElement(TriggersTag)
	}
	// TODO issue: if support RemoteTrigger in console, need to delete GenericWebhook tag when pipeline.GenericWebhook is nil;
	if pipeline.RemoteTrigger != nil {
		addOrUpdateElement(flow, AuthTokenTag, pipeline.RemoteTrigger.Token)
	}
	addOrUpdateElement(flow, DisabledTag, "false")

	// format xml string
	doc.Indent(2)
	stringXml, err := doc.WriteToString()
	if err != nil {
		return "", err
	}
	return replaceXmlVersion(stringXml, "1.0", "1.1"), err
}

func parsePipelineConfigXml(config string) (*devopsv1alpha3.NoScmPipeline, error) {
	pipeline := &devopsv1alpha3.NoScmPipeline{}
	config = replaceXmlVersion(config, "1.1", "1.0")
	doc := etree.NewDocument()
	err := doc.ReadFromString(config)
	if err != nil {
		return nil, err
	}
	flow := doc.SelectElement("flow-definition")
	if flow == nil {
		return nil, fmt.Errorf("can not find pipeline definition")
	}
	if node := flow.SelectElement("description"); node != nil {
		pipeline.Description = node.Text()
	}

	properties := flow.SelectElement("properties")
	if properties.
		SelectElement(
			"org.jenkinsci.plugins.workflow.job.properties.DisableConcurrentBuildsJobProperty") != nil {
		pipeline.DisableConcurrent = true
	}
	if properties.SelectElement("jenkins.model.BuildDiscarderProperty") != nil {
		strategy := properties.
			SelectElement("jenkins.model.BuildDiscarderProperty").
			SelectElement("strategy")
		pipeline.Discarder = &devopsv1alpha3.DiscarderProperty{
			DaysToKeep: getElementTextValueOrEmpty(strategy, "daysToKeep"),
			NumToKeep:  getElementTextValueOrEmpty(strategy, "numToKeep"),
		}
	}

	pipeline.Parameters = getParametersfromEtree(properties)
	if len(pipeline.Parameters) == 0 {
		pipeline.Parameters = nil
	}

	if triggerProperty := properties.
		SelectElement(
			"org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty"); triggerProperty != nil {
		triggersEle := triggerProperty.SelectElement("triggers")
		if timerTrigger := triggersEle.SelectElement("hudson.triggers.TimerTrigger"); timerTrigger != nil {
			pipeline.TimerTrigger = &devopsv1alpha3.TimerTrigger{
				Cron: getElementTextValueOrEmpty(timerTrigger, "spec"),
			}
		}

		if genericWebhookEle := triggersEle.SelectElement("org.jenkinsci.plugins.gwt.GenericTrigger"); genericWebhookEle != nil {
			pipeline.GenericWebhook = triggers.ParseGenericWebhookXML(genericWebhookEle)
		} else if pipeline.GenericWebhook != nil {
			pipeline.GenericWebhook.Enable = false
		}
	}
	if authToken := flow.SelectElement("authToken"); authToken != nil {
		pipeline.RemoteTrigger = &devopsv1alpha3.RemoteTrigger{
			Token: authToken.Text(),
		}
	}
	if definition := flow.SelectElement("definition"); definition != nil {
		if script := definition.SelectElement("script"); script != nil {
			pipeline.Jenkinsfile = script.Text()
		}
	}
	return pipeline, nil
}

func replaceParametersInEtree(properties *etree.Element, parameters []devopsv1alpha3.ParameterDefinition) {
	var paramDefiPropsE, paramDefiE *etree.Element
	if paramDefiPropsE = properties.SelectElement(ParamDefiPropTag); paramDefiPropsE == nil {
		paramDefiE = properties.CreateElement(ParamDefiPropTag).CreateElement(ParamDefiTag)
	} else {
		if paramDefiE = paramDefiPropsE.SelectElement(ParamDefiTag); paramDefiE != nil {
			paramDefiPropsE.RemoveChild(paramDefiE)
		}
		paramDefiE = paramDefiPropsE.CreateElement(ParamDefiTag)
	}

	for _, parameter := range parameters {
		for className, typeName := range ParameterTypeMap {
			if typeName == parameter.Type {
				paramDefine := paramDefiE.CreateElement(className)
				paramDefine.CreateElement("name").SetText(parameter.Name)
				paramDefine.CreateElement("description").SetText(parameter.Description)
				switch parameter.Type {
				case "choice":
					choices := paramDefine.CreateElement("choices")
					choices.CreateAttr("class", "java.util.Arrays$ArrayList")
					// see also https://github.com/kubesphere/kubesphere/issues/3430
					a := choices.CreateElement("a")
					a.CreateAttr("class", "string-array")
					choiceValues := strings.Split(parameter.DefaultValue, "\n")
					for _, choiceValue := range choiceValues {
						a.CreateElement("string").SetText(choiceValue)
					}
				case "file":
					break
				default:
					paramDefine.CreateElement("defaultValue").SetText(parameter.DefaultValue)
				}
			}
		}
	}
}

func getElementTextValueOrEmpty(element *etree.Element, name string) string {
	subEle := element.SelectElement(name)
	if subEle != nil {
		return subEle.Text()
	}
	return ""
}

func getParametersfromEtree(properties *etree.Element) []devopsv1alpha3.ParameterDefinition {
	var parameters []devopsv1alpha3.ParameterDefinition
	if parametersProperty := properties.SelectElement("hudson.model.ParametersDefinitionProperty"); parametersProperty != nil {
		params := parametersProperty.SelectElement("parameterDefinitions").ChildElements()
		for _, param := range params {
			switch param.Tag {
			case "hudson.model.StringParameterDefinition":
				parameters = append(parameters, devopsv1alpha3.ParameterDefinition{
					Name:         getElementTextValueOrEmpty(param, "name"),
					Description:  getElementTextValueOrEmpty(param, "description"),
					DefaultValue: getElementTextValueOrEmpty(param, "defaultValue"),
					Type:         ParameterTypeMap["hudson.model.StringParameterDefinition"],
				})
			case "hudson.model.BooleanParameterDefinition":
				parameters = append(parameters, devopsv1alpha3.ParameterDefinition{
					Name:         getElementTextValueOrEmpty(param, "name"),
					Description:  getElementTextValueOrEmpty(param, "description"),
					DefaultValue: getElementTextValueOrEmpty(param, "defaultValue"),
					Type:         ParameterTypeMap["hudson.model.BooleanParameterDefinition"],
				})
			case "hudson.model.TextParameterDefinition":
				parameters = append(parameters, devopsv1alpha3.ParameterDefinition{
					Name:         getElementTextValueOrEmpty(param, "name"),
					Description:  getElementTextValueOrEmpty(param, "description"),
					DefaultValue: getElementTextValueOrEmpty(param, "defaultValue"),
					Type:         ParameterTypeMap["hudson.model.TextParameterDefinition"],
				})
			case "hudson.model.FileParameterDefinition":
				parameters = append(parameters, devopsv1alpha3.ParameterDefinition{
					Name:        getElementTextValueOrEmpty(param, "name"),
					Description: getElementTextValueOrEmpty(param, "description"),
					Type:        ParameterTypeMap["hudson.model.FileParameterDefinition"],
				})
			case "hudson.model.PasswordParameterDefinition":
				parameters = append(parameters, devopsv1alpha3.ParameterDefinition{
					Name:         getElementTextValueOrEmpty(param, "name"),
					Description:  getElementTextValueOrEmpty(param, "description"),
					DefaultValue: getElementTextValueOrEmpty(param, "name"),
					Type:         ParameterTypeMap["hudson.model.PasswordParameterDefinition"],
				})
			case "hudson.model.ChoiceParameterDefinition":
				choiceParameter := devopsv1alpha3.ParameterDefinition{
					Name:        getElementTextValueOrEmpty(param, "name"),
					Description: getElementTextValueOrEmpty(param, "description"),
					Type:        ParameterTypeMap["hudson.model.ChoiceParameterDefinition"],
				}
				choicesEle := param.SelectElement("choices")
				var choices []*etree.Element
				// the child element is a in the simple pipeline, the child is string list in the multi-branch pipeline
				// see also https://github.com/kubesphere/kubesphere/issues/3430
				choiceAnchor := choicesEle.SelectElement("a")
				if choiceAnchor == nil {
					choices = choicesEle.SelectElements("string")
				} else {
					choices = choiceAnchor.SelectElements("string")
				}
				for _, choice := range choices {
					choiceParameter.DefaultValue += fmt.Sprintf("%s\n", choice.Text())
				}
				choiceParameter.DefaultValue = strings.TrimSpace(choiceParameter.DefaultValue)
				parameters = append(parameters, choiceParameter)
			default:
				parameters = append(parameters, devopsv1alpha3.ParameterDefinition{
					Name:         getElementTextValueOrEmpty(param, "name"),
					Description:  getElementTextValueOrEmpty(param, "description"),
					DefaultValue: "unknown",
					Type:         param.Tag,
				})
			}
		}
	}
	return parameters
}

func appendMultiBranchJobTriggerToEtree(properties *etree.Element, s *devopsv1alpha3.MultiBranchJobTrigger) {
	triggerProperty := properties.CreateElement("org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty")
	triggerProperty.CreateAttr("plugin", "multibranch-action-triggers")
	triggerProperty.CreateElement("createActionJobsToTrigger").SetText(s.CreateActionJobsToTrigger)
	triggerProperty.CreateElement("deleteActionJobsToTrigger").SetText(s.DeleteActionJobsToTrigger)
	return
}

func getMultiBranchJobTriggerfromEtree(properties *etree.Element) *devopsv1alpha3.MultiBranchJobTrigger {
	var s devopsv1alpha3.MultiBranchJobTrigger
	triggerProperty := properties.SelectElement("org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty")
	if triggerProperty != nil {
		s.CreateActionJobsToTrigger = getElementTextValueOrEmpty(triggerProperty, "createActionJobsToTrigger")
		s.DeleteActionJobsToTrigger = getElementTextValueOrEmpty(triggerProperty, "deleteActionJobsToTrigger")
	}
	return &s
}
func createMultiBranchPipelineConfigXml(projectName string, pipeline *devopsv1alpha3.MultiBranchPipeline) (string, error) {
	doc := etree.NewDocument()
	xmlString := `
<?xml version='1.0' encoding='UTF-8'?>
<org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject plugin="workflow-multibranch">
  <actions/>
  <properties>
    <org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig plugin="pipeline-model-definition">
      <dockerLabel></dockerLabel>
      <registry plugin="docker-commons"/>
    </org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig>
  </properties>
  <folderViews class="jenkins.branch.MultiBranchProjectViewHolder" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </folderViews>
  <healthMetrics>
    <com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric plugin="cloudbees-folder">
      <nonRecursive>false</nonRecursive>
    </com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
  </healthMetrics>
  <icon class="jenkins.branch.MetadataActionFolderIcon" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </icon>
</org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject>`
	err := doc.ReadFromString(xmlString)
	if err != nil {
		return "", err
	}

	project := doc.SelectElement("org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject")
	project.CreateElement("description").SetText(pipeline.Description)

	if pipeline.MultiBranchJobTrigger != nil {
		properties := project.SelectElement("properties")
		appendMultiBranchJobTriggerToEtree(properties, pipeline.MultiBranchJobTrigger)
	}

	discarder := project.CreateElement("orphanedItemStrategy")
	discarder.CreateAttr("class", "com.cloudbees.hudson.plugins.folder.computed.DefaultOrphanedItemStrategy")
	discarder.CreateAttr("plugin", "cloudbees-folder")
	if pipeline.Discarder != nil {
		discarder.CreateElement("pruneDeadBranches").SetText("true")
		discarder.CreateElement("daysToKeep").SetText(pipeline.Discarder.DaysToKeep)
		discarder.CreateElement("numToKeep").SetText(pipeline.Discarder.NumToKeep)
	} else {
		discarder.CreateElement("pruneDeadBranches").SetText("false")
	}

	triggers := project.CreateElement("triggers")
	if pipeline.TimerTrigger != nil {
		timeTrigger := triggers.CreateElement(
			"com.cloudbees.hudson.plugins.folder.computed.PeriodicFolderTrigger")
		timeTrigger.CreateAttr("plugin", "cloudbees-folder")
		millis, err := strconv.ParseInt(pipeline.TimerTrigger.Interval, 10, 64)
		if err != nil {
			return "", err
		}
		timeTrigger.CreateElement("spec").SetText(toCrontab(millis))
		timeTrigger.CreateElement("interval").SetText(pipeline.TimerTrigger.Interval)

		triggers.CreateElement("disabled").SetText("false")
	}

	sources := project.CreateElement("sources")
	sources.CreateAttr("class", "jenkins.branch.MultiBranchProject$BranchSourceList")
	sources.CreateAttr("plugin", "branch-api")
	sourcesOwner := sources.CreateElement("owner")
	sourcesOwner.CreateAttr("class", "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject")
	sourcesOwner.CreateAttr("reference", "../..")

	branchSource := sources.CreateElement("data").CreateElement("jenkins.branch.BranchSource")
	branchSourceStrategy := branchSource.CreateElement("strategy")
	branchSourceStrategy.CreateAttr("class", "jenkins.branch.NamedExceptionsBranchPropertyStrategy")
	branchSourceStrategy.CreateElement("defaultProperties").CreateAttr("class", "empty-list")
	branchSourceStrategy.CreateElement("namedExceptions").CreateAttr("class", "empty-list")
	source := branchSource.CreateElement("source")

	switch pipeline.SourceType {
	case devopsv1alpha3.SourceTypeGit:
		internal.AppendGitSourceToEtree(source, pipeline.GitSource)
	case devopsv1alpha3.SourceTypeGithub:
		internal.AppendGithubSourceToEtree(source, pipeline.GitHubSource)
	case devopsv1alpha3.SourceTypeGitlab:
		internal.AppendGitlabSourceToEtree(source, pipeline.GitlabSource)
	case devopsv1alpha3.SourceTypeSVN:
		internal.AppendSvnSourceToEtree(source, pipeline.SvnSource)
	case devopsv1alpha3.SourceTypeSingleSVN:
		internal.AppendSingleSvnSourceToEtree(source, pipeline.SingleSvnSource)
	case devopsv1alpha3.SourceTypeBitbucket:
		internal.AppendBitbucketServerSourceToEtree(source, pipeline.BitbucketServerSource)

	default:
		return "", fmt.Errorf("unsupport source type: %s", pipeline.SourceType)
	}

	factory := project.CreateElement("factory")
	factory.CreateAttr("class", "org.jenkinsci.plugins.workflow.multibranch.WorkflowBranchProjectFactory")
	factoryOwner := factory.CreateElement("owner")
	factoryOwner.CreateAttr("class", "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject")
	factoryOwner.CreateAttr("reference", "../..")
	factory.CreateElement("scriptPath").SetText(pipeline.ScriptPath)

	doc.Indent(2)
	stringXml, err := doc.WriteToString()
	return replaceXmlVersion(stringXml, "1.0", "1.1"), err
}

func parseMultiBranchPipelineConfigXml(config string) (*devopsv1alpha3.MultiBranchPipeline, error) {
	pipeline := &devopsv1alpha3.MultiBranchPipeline{}
	config = replaceXmlVersion(config, "1.1", "1.0")
	doc := etree.NewDocument()
	err := doc.ReadFromString(config)
	if err != nil {
		return nil, err
	}
	project := doc.SelectElement("org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject")
	if project == nil {
		return nil, fmt.Errorf("can not parse mutibranch pipeline config")
	}
	if properties := project.SelectElement("properties"); properties != nil {
		if multibranchTrigger := properties.SelectElement(
			"org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty"); multibranchTrigger != nil {
			pipeline.MultiBranchJobTrigger = getMultiBranchJobTriggerfromEtree(properties)
		}
	}
	if project.SelectElement("description") != nil {
		pipeline.Description = getElementTextValueOrEmpty(project, "description")
	}

	if discarder := project.SelectElement("orphanedItemStrategy"); discarder != nil {
		if getElementTextValueOrEmpty(discarder, "pruneDeadBranches") == "true" {
			pipeline.Discarder = &devopsv1alpha3.DiscarderProperty{
				DaysToKeep: getElementTextValueOrEmpty(discarder, "daysToKeep"),
				NumToKeep:  getElementTextValueOrEmpty(discarder, "numToKeep"),
			}
		}
	}
	if triggers := project.SelectElement("triggers"); triggers != nil {
		if timerTrigger := triggers.SelectElement(
			"com.cloudbees.hudson.plugins.folder.computed.PeriodicFolderTrigger"); timerTrigger != nil {
			pipeline.TimerTrigger = &devopsv1alpha3.TimerTrigger{
				Interval: getElementTextValueOrEmpty(timerTrigger, "interval"),
			}
		}
	}

	if sources := project.SelectElement("sources"); sources != nil {
		if sourcesData := sources.SelectElement("data"); sourcesData != nil {
			if branchSource := sourcesData.SelectElement("jenkins.branch.BranchSource"); branchSource != nil {
				source := branchSource.SelectElement("source")
				switch source.SelectAttr("class").Value {
				case "org.jenkinsci.plugins.github_branch_source.GitHubSCMSource":
					pipeline.GitHubSource = internal.GetGithubSourcefromEtree(source)
					pipeline.SourceType = devopsv1alpha3.SourceTypeGithub
				case "com.cloudbees.jenkins.plugins.bitbucket.BitbucketSCMSource":
					pipeline.BitbucketServerSource = internal.GetBitbucketServerSourceFromEtree(source)
					pipeline.SourceType = devopsv1alpha3.SourceTypeBitbucket
				case "io.jenkins.plugins.gitlabbranchsource.GitLabSCMSource":
					pipeline.GitlabSource = internal.GetGitlabSourceFromEtree(source)
					pipeline.SourceType = devopsv1alpha3.SourceTypeGitlab

				case "jenkins.plugins.git.GitSCMSource":
					pipeline.SourceType = devopsv1alpha3.SourceTypeGit
					pipeline.GitSource = internal.GetGitSourcefromEtree(source)

				case "jenkins.scm.impl.SingleSCMSource":
					pipeline.SourceType = devopsv1alpha3.SourceTypeSingleSVN
					pipeline.SingleSvnSource = internal.GetSingleSvnSourceFromEtree(source)

				case "jenkins.scm.impl.subversion.SubversionSCMSource":
					pipeline.SourceType = devopsv1alpha3.SourceTypeSVN
					pipeline.SvnSource = internal.GetSvnSourcefromEtree(source)
				}
			}
		}
	}

	scriptPathEle := project.SelectElement("factory").SelectElement("scriptPath")
	if scriptPathEle != nil {
		// There's no script path if current pipeline using a default Jenkinsfile
		// see also https://github.com/jenkinsci/pipeline-multibranch-defaults-plugin
		pipeline.ScriptPath = scriptPathEle.Text()
	}
	return pipeline, nil
}

func toCrontab(millis int64) string {
	if millis*time.Millisecond.Nanoseconds() <= 5*time.Minute.Nanoseconds() {
		return "* * * * *"
	}
	if millis*time.Millisecond.Nanoseconds() <= 30*time.Minute.Nanoseconds() {
		return "H/5 * * * *"
	}
	if millis*time.Millisecond.Nanoseconds() <= 1*time.Hour.Nanoseconds() {
		return "H/15 * * * *"
	}
	if millis*time.Millisecond.Nanoseconds() <= 8*time.Hour.Nanoseconds() {
		return "H/30 * * * *"
	}
	if millis*time.Millisecond.Nanoseconds() <= 24*time.Hour.Nanoseconds() {
		return "H H/4 * * *"
	}
	if millis*time.Millisecond.Nanoseconds() <= 48*time.Hour.Nanoseconds() {
		return "H H/12 * * *"
	}
	return "H H * * *"

}

func addOrUpdateElement(parent *etree.Element, tag, text string) *etree.Element {
	var e *etree.Element
	if e = parent.SelectElement(tag); e == nil {
		e = parent.CreateElement(tag)
	}
	if text != "" {
		e.SetText(text)
	}
	return e
}

func replaceAttr(e *etree.Element, key, value string) *etree.Element {
	e.RemoveAttr(key)
	e.CreateAttr(key, value)
	return e
}

func removeChildElement(parent *etree.Element, childTag string) *etree.Element {
	if e := parent.SelectElement(childTag); e != nil {
		parent.RemoveChild(e)
	}
	return parent
}
