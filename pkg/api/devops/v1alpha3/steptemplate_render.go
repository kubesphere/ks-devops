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

package v1alpha3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	v1 "k8s.io/api/core/v1"
)

// Render renders the template and returns the result
func (t *StepTemplateSpec) Render(param map[string]interface{}, secret *v1.Secret) (output string, err error) {
	// taking the default parameter values
	for i := range t.Parameters {
		item := t.Parameters[i]
		if _, ok := param[item.Name]; !ok && item.DefaultValue != "" {
			param[item.Name] = item.DefaultValue
		}
	}

	switch t.Runtime {
	case "dsl":
		output, err = dslRender(t.Template, param, secret)
	case "shell":
		fallthrough
	default:
		output, err = shellRender(t.Template, param, secret)
	}

	if t.Secret.Wrap && secret != nil {
		output = wrapWithCredential(t.Secret.Type, secret.Name, output)
	}

	if err == nil && t.Container != "" {
		output = fmt.Sprintf(`
{
  "arguments": {
	"isLiteral": true,
	"value": "%s"
  },
  "children": [%s],
  "name": "container"
}`, t.Container, output)
	}

	output = jsonFormat(output)
	return
}

func jsonFormat(content string) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(content), "", "  "); err == nil {
		return strings.TrimSpace(buf.String())
	}
	return strings.TrimSpace(content)
}

func wrapWithCredential(secretType, secretName, target string) string {
	switch secretType {
	case string(v1.SecretTypeBasicAuth), string(SecretTypeBasicAuth):
		target = fmt.Sprintf(`{
      "arguments": {
        "isLiteral": false,
        "value": "${[usernamePassword(credentialsId: '%s', passwordVariable: 'PASSWORDVARIABLE' ,usernameVariable : 'USERNAMEVARIABLE')]}"
      },
      "children": [%s],
      "name": "withCredentials"
    }`, secretName, target)
	case string(v1.SecretTypeBootstrapToken), string(SecretTypeSecretText):
		target = fmt.Sprintf(`{
  "arguments": {
    "isLiteral": false,
    "value": "${[string(credentialsId: '%s', variable: 'VARIABLE')]}"
  },
  "children": [%s],
  "name": "withCredentials"
}`, secretName, target)
	case string(SecretTypeKubeConfig):
		target = fmt.Sprintf(`{
  "arguments": {
    "isLiteral": false,
    "value": "${[kubeconfigContent(credentialsId: '%s', variable: 'VARIABLE')]}"
  },
  "children": [%s],
  "name": "withCredentials"
}`, secretName, target)
	case string(SecretTypeSSHAuth):
		target = fmt.Sprintf(`{
  "arguments": {
    "isLiteral": false,
    "value": "${[sshUserPrivateKey(credentialsId: '%s', keyFileVariable : 'KEYFILEVARIABLE' ,passphraseVariable : 'PASSPHRASEVARIABLE' ,usernameVariable : 'SSHUSERPRIVATEKEY')]}"
  },
  "children": [%s],
  "name": "withCredentials"
}`, secretName, target)
	}
	return jsonFormat(target)
}

func dslRender(dslTpl string, param map[string]interface{}, secret *v1.Secret) (output string, err error) {
	output = dslTpl

	var tpl *template.Template
	if tpl, err = template.New("shell").Parse(output); err != nil {
		return
	}

	data := map[string]interface{}{
		"param":  param,
		"secret": secret,
	}

	buf := bytes.NewBuffer([]byte{})
	if err = tpl.Execute(buf, data); err == nil {
		output = strings.TrimSpace(buf.String())
	}
	return
}

func shellRender(shellTpl string, param map[string]interface{}, secret *v1.Secret) (output string, err error) {
	if output, err = dslRender(shellTpl, param, secret); err == nil {
		escapedOutput := strings.ReplaceAll(output, `\`, `\\`)
		escapedOutput = strings.ReplaceAll(escapedOutput, "\n", "\\n")
		escapedOutput = strings.ReplaceAll(escapedOutput, `"`, `\"`)

		output = fmt.Sprintf(`{
"arguments": [
  {
	"key": "script",
	"value": {
	  "isLiteral": true,
	  "value": "%s"
	}
  }
],
"name": "sh"
}`, escapedOutput)
	}
	return
}
