/*
Copyright 2022 The KubeSphere Authors.

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

package triggers

import (
	"fmt"
	"github.com/beevik/etree"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"strconv"
)

// CreateGenericWebhookXML creates the xml element for GenericTrigger
func CreateGenericWebhookXML(parent *etree.Element, webhook *v1alpha3.GenericWebhook) (ele *etree.Element) {
	if webhook == nil || parent == nil || !webhook.Enable {
		return
	}

	ele = parent.CreateElement("org.jenkinsci.plugins.gwt.GenericTrigger")

	ele.CreateElement("spec")
	ele.CreateElement("token").SetText(webhook.Token)
	ele.CreateElement("causeString").SetText(webhook.Cause)
	ele.CreateElement("printContributedVariables").SetText(fmt.Sprintf("%v", webhook.PrintVariables))
	ele.CreateElement("printPostContent").SetText(fmt.Sprintf("%v", webhook.PrintPostContent))
	ele.CreateElement("regexpFilterText").SetText(webhook.FilterText)
	ele.CreateElement("regexpFilterExpression").SetText(webhook.FilterExpression)

	requestVarsEle := ele.CreateElement("genericRequestVariables")
	for _, item := range webhook.RequestVariables {
		varEle := requestVarsEle.CreateElement("org.jenkinsci.plugins.gwt.GenericRequestVariable")
		varEle.CreateElement("key").SetText(item.Key)
		varEle.CreateElement("regexpFilter").SetText(item.RegexpFilter)
	}

	headerVarsEle := ele.CreateElement("genericHeaderVariables")
	for _, item := range webhook.HeaderVariables {
		varEle := headerVarsEle.CreateElement("org.jenkinsci.plugins.gwt.GenericHeaderVariable")
		varEle.CreateElement("key").SetText(item.Key)
		varEle.CreateElement("regexpFilter").SetText(item.RegexpFilter)
	}
	return
}

// ParseGenericWebhookXML parse GenericTrigger xml structure into go struct GenericWebhook
func ParseGenericWebhookXML(ele *etree.Element) (webhook *v1alpha3.GenericWebhook) {
	if ele == nil {
		return
	}

	webhook = &v1alpha3.GenericWebhook{
		Enable:           true,
		Token:            getElementText(ele, "token"),
		Cause:            getElementText(ele, "causeString"),
		PrintVariables:   getElementTextAsBoolean(ele, "printContributedVariables"),
		PrintPostContent: getElementTextAsBoolean(ele, "printPostContent"),
		FilterText:       getElementText(ele, "regexpFilterText"),
		FilterExpression: getElementText(ele, "regexpFilterExpression"),
	}

	if reqVarsEle := ele.SelectElement("genericRequestVariables"); reqVarsEle != nil {
		if eles := reqVarsEle.SelectElements("org.jenkinsci.plugins.gwt.GenericRequestVariable"); eles != nil {
			webhook.RequestVariables = []v1alpha3.GenericVariable{}

			for i, _ := range eles {
				webhook.RequestVariables = append(webhook.RequestVariables, v1alpha3.GenericVariable{
					Key:          getElementText(eles[i], "key"),
					RegexpFilter: getElementText(eles[i], "regexpFilter"),
				})
			}
		}
	}

	if reqVarsEle := ele.SelectElement("genericHeaderVariables"); reqVarsEle != nil {
		if eles := reqVarsEle.SelectElements("org.jenkinsci.plugins.gwt.GenericHeaderVariable"); eles != nil {
			webhook.HeaderVariables = []v1alpha3.GenericVariable{}

			for i, _ := range eles {
				webhook.HeaderVariables = append(webhook.HeaderVariables, v1alpha3.GenericVariable{
					Key:          getElementText(eles[i], "key"),
					RegexpFilter: getElementText(eles[i], "regexpFilter"),
				})
			}
		}
	}
	return
}

func getElementText(ele *etree.Element, childName string) string {
	if ele == nil {
		return ""
	}

	if child := ele.SelectElement(childName); child != nil {
		return child.Text()
	}
	return ""
}

func getElementTextAsBoolean(ele *etree.Element, childName string) (result bool) {
	if strVal := getElementText(ele, childName); strVal != "" {
		result, _ = strconv.ParseBool(strVal)
	}
	return
}
