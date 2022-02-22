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
	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"reflect"
	"strconv"
	"testing"
)

func TestCreateGenericWebhookXML(t *testing.T) {
	ele := CreateGenericWebhookXML(nil, nil)
	assert.Nil(t, ele)

	webhook := v1alpha3.GenericWebhook{
		Enable:           true,
		Token:            "token",
		Cause:            "cause",
		PrintPostContent: true,
		PrintVariables:   false,
		FilterText:       "filterText",
		FilterExpression: "filterExpression",
		HeaderVariables: []v1alpha3.GenericVariable{{
			Key:          "key",
			RegexpFilter: "regexpFilter",
		}},
		RequestVariables: []v1alpha3.GenericVariable{{
			Key:          "key",
			RegexpFilter: "regexpFilter",
		}},
	}
	ele = CreateGenericWebhookXML(&etree.Element{}, &webhook)
	assert.NotNil(t, ele.SelectElement("spec"), "the element spec is mandatory")
	assert.Equal(t, ele.Tag, "org.jenkinsci.plugins.gwt.GenericTrigger")
	assert.Equal(t, ele.SelectElement("token").Text(), webhook.Token)
	assert.Equal(t, ele.SelectElement("causeString").Text(), webhook.Cause)
	assert.Equal(t, ele.SelectElement("printContributedVariables").Text(), strconv.FormatBool(webhook.PrintVariables))
	assert.Equal(t, ele.SelectElement("printPostContent").Text(), strconv.FormatBool(webhook.PrintPostContent))
	assert.Equal(t, ele.SelectElement("regexpFilterText").Text(), webhook.FilterText)
	assert.Equal(t, ele.SelectElement("regexpFilterExpression").Text(), webhook.FilterExpression)

	requestVarsEle := ele.SelectElement("genericRequestVariables").SelectElements("org.jenkinsci.plugins.gwt.GenericRequestVariable")
	assert.Equal(t, len(requestVarsEle), len(webhook.RequestVariables))
	assert.Equal(t, requestVarsEle[0].SelectElement("key").Text(), webhook.RequestVariables[0].Key)
	assert.Equal(t, requestVarsEle[0].SelectElement("regexpFilter").Text(), webhook.RequestVariables[0].RegexpFilter)

	headerVarsEle := ele.SelectElement("genericHeaderVariables").SelectElements("org.jenkinsci.plugins.gwt.GenericHeaderVariable")
	assert.Equal(t, len(headerVarsEle), len(webhook.HeaderVariables))
	assert.Equal(t, headerVarsEle[0].SelectElement("key").Text(), webhook.HeaderVariables[0].Key)
	assert.Equal(t, headerVarsEle[0].SelectElement("regexpFilter").Text(), webhook.HeaderVariables[0].RegexpFilter)
}

func TestParseGenericWebhookXML(t *testing.T) {
	type args struct {
		ele *etree.Element
	}
	tests := []struct {
		name        string
		args        args
		wantWebhook *v1alpha3.GenericWebhook
	}{{
		name: "nil args, should return nil",
	}, {
		name: "normal case",
		args: args{
			ele: &etree.Element{
				Tag: "org.jenkinsci.plugins.gwt.GenericTrigger",
				Child: []etree.Token{&etree.Element{
					Tag: "token",
					Child: []etree.Token{&etree.CharData{
						Data: "token",
					}},
				}, &etree.Element{
					Tag: "printPostContent",
					Child: []etree.Token{&etree.CharData{
						Data: "true",
					}},
				}, &etree.Element{
					Tag: "printContributedVariables",
					Child: []etree.Token{&etree.CharData{
						Data: "true",
					}},
				}, &etree.Element{
					Tag: "causeString",
					Child: []etree.Token{&etree.CharData{
						Data: "cause",
					}},
				}, &etree.Element{
					Tag: "regexpFilterText",
					Child: []etree.Token{&etree.CharData{
						Data: "filterText",
					}},
				}, &etree.Element{
					Tag: "regexpFilterExpression",
					Child: []etree.Token{&etree.CharData{
						Data: "filterExpression",
					}},
				}, &etree.Element{
					Tag: "genericRequestVariables",
					Child: []etree.Token{&etree.Element{
						Tag: "org.jenkinsci.plugins.gwt.GenericRequestVariable",
						Child: []etree.Token{&etree.Element{
							Tag: "key",
							Child: []etree.Token{&etree.CharData{
								Data: "key",
							}},
						}, &etree.Element{
							Tag: "regexpFilter",
							Child: []etree.Token{&etree.CharData{
								Data: "regexpFilter",
							}},
						}},
					}},
				}, &etree.Element{
					Tag: "genericHeaderVariables",
					Child: []etree.Token{&etree.Element{
						Tag: "org.jenkinsci.plugins.gwt.GenericHeaderVariable",
						Child: []etree.Token{&etree.Element{
							Tag: "key",
							Child: []etree.Token{&etree.CharData{
								Data: "key",
							}},
						}, &etree.Element{
							Tag: "regexpFilter",
							Child: []etree.Token{&etree.CharData{
								Data: "regexpFilter",
							}},
						}},
					}},
				}},
			}},
		wantWebhook: &v1alpha3.GenericWebhook{
			Enable:           true,
			Token:            "token",
			Cause:            "cause",
			PrintVariables:   true,
			PrintPostContent: true,
			RequestVariables: []v1alpha3.GenericVariable{{
				Key:          "key",
				RegexpFilter: "regexpFilter",
			}},
			HeaderVariables: []v1alpha3.GenericVariable{{
				Key:          "key",
				RegexpFilter: "regexpFilter",
			}},
			FilterText:       "filterText",
			FilterExpression: "filterExpression",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotWebhook := ParseGenericWebhookXML(tt.args.ele); !reflect.DeepEqual(gotWebhook, tt.wantWebhook) {
				t.Errorf("ParseGenericWebhookXML() = %v, want %v", gotWebhook, tt.wantWebhook)
			}
		})
	}
}

func Test_getElementText(t *testing.T) {
	type args struct {
		ele       *etree.Element
		childName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "nil args, should return empty string",
		args: args{
			ele: nil,
		},
	}, {
		name: "non-exits child name, should return empty string",
		args: args{
			ele:       etree.NewElement("xxx"),
			childName: "non-exits",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getElementText(tt.args.ele, tt.args.childName); got != tt.want {
				t.Errorf("getElementText() = %v, want %v", got, tt.want)
			}
		})
	}
}
