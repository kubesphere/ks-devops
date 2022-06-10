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
	"fmt"
	v1 "k8s.io/api/core/v1"
	"strings"
	"text/template"
)

func (t *StepTemplateSpec) Render(param map[string]string, secret *v1.Secret) (output string, err error) {
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

	if t.Secret.Wrap {
		output = wrapWithCredential(t.Secret.Type, secret.Name, output)
	}

	if err == nil && t.Container != "" {
		output = fmt.Sprintf(`container("%s") {
%s
}`, t.Container, addIndent(output))
	}
	return
}

func addIndent(txt string) string {
	items := strings.Split(txt, "\n")
	return "\t" + strings.Join(items, "\n\t")
}

func wrapWithCredential(secretType, secretName, target string) string {
	switch secretType {
	case string(v1.SecretTypeBasicAuth), string(SecretTypeBasicAuth):
		return fmt.Sprintf(`withCredential[usernamePassword(credentialsId : "%s" ,passwordVariable : 'PASSWORDVARIABLE' ,usernameVariable : 'USERNAMEVARIABLE')]) {
%s
}`, secretName, addIndent(target))
	case string(v1.SecretTypeBootstrapToken), string(SecretTypeSecretText):
		return fmt.Sprintf(`withCredential[string(credentialsId : "%s" ,variable : 'VARIABLE')]) {
%s
}`, secretName, addIndent(target))
	case string(SecretTypeKubeConfig):
		return fmt.Sprintf(`withCredential[kubeconfigContent(credentialsId : "%s" ,variable : 'VARIABLE')]) {
%s
}`, secretName, addIndent(target))
	case string(SecretTypeSSHAuth):
		return fmt.Sprintf(`withCredential[sshUserPrivateKey(credentialsId : "%s" ,keyFileVariable : 'KEYFILEVARIABLE' ,passphraseVariable : 'PASSPHRASEVARIABLE' ,usernameVariable : 'SSHUSERPRIVATEKEY')]) {
%s
}`, secretName, addIndent(target))
	}
	return target
}

func dslRender(dslTpl string, param map[string]string, secret *v1.Secret) (output string, err error) {
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

func shellRender(shellTpl string, param map[string]string, secret *v1.Secret) (output string, err error) {
	if output, err = dslRender(shellTpl, param, secret); err == nil {
		output = fmt.Sprintf(`sh '''
%s
'''`, addIndent(output))
	}
	return
}
