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
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_handler_stepTemplateRender(t *testing.T) {
	stepTemplate := &StepTemplateSpec{
		Template: `cat > log.sh << EOF
        for ((i=1; i<=1000000; i++))
        do
            echo "Log message number \\$i: This is a sample log message."
        done
EOF

        cat log.sh
        bash log.sh`,
		Runtime: "shell",
	}
	stepTemplateWithParameters := &StepTemplateSpec{
		Template: `docker login -u $USERNAMEVARIABLE -p $PASSWORDVARIABLE
docker build {{.param.context}} -t {{.param.tag}} -f {{.param.dockerfile.path}}`,
		Runtime:   "shell",
		Container: "base",
		Secret: SecretInStep{
			Wrap: true,
			Type: string(v1.SecretTypeBasicAuth),
		},
		Parameters: []ParameterInStep{{
			Name: "context",
		}, {
			Name: "tag",
		}, {
			Name: "dockerfile",
		}},
	}
	stepTemplateWithDSL := &StepTemplateSpec{
		Template: readFile("testdata/dsl-echo.json"),
		Runtime:  "dsl",
	}

	type args struct {
		stepTemplate *StepTemplateSpec
		param        map[string]interface{}
		secret       *v1.Secret
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantErr    bool
	}{{
		name: "a simple shell without any parameters",
		args: args{
			stepTemplate: stepTemplate,
		},
		wantOutput: `{
  "arguments": [
    {
      "key": "script",
      "value": {
        "isLiteral": true,
        "value": "cat > log.sh << EOF\n        for ((i=1; i<=1000000; i++))\n        do\n            echo \"Log message number \\\\$i: This is a sample log message.\"\n        done\nEOF\n\n        cat log.sh\n        bash log.sh"
      }
    }
  ],
  "name": "sh"
}`,
		wantErr: false,
	}, {
		name: "docker build command with parameters",
		args: args{
			stepTemplate: stepTemplateWithParameters,
			param: map[string]interface{}{
				"context": "dir",
				"tag":     "image:tag",
				"dockerfile": map[string]string{
					"path": "Dockerfile",
				},
			},
			secret: &v1.Secret{
				ObjectMeta: v12.ObjectMeta{
					Name: "docker",
				},
				Type: v1.SecretTypeBasicAuth,
				Data: map[string][]byte{
					v1.BasicAuthUsernameKey: []byte("username"),
					v1.BasicAuthPasswordKey: []byte("password"),
				},
			},
		},
		wantOutput: readFile("testdata/docker-login.json"),
		wantErr:    false,
	}, {
		name: "a simple dsl without any parameters",
		args: args{
			stepTemplate: stepTemplateWithDSL,
		},
		wantOutput: readFile("testdata/dsl-echo.json"),
		wantErr:    false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutput, err := tt.args.stepTemplate.Render(tt.args.param, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("stepTemplateRender() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOutput != tt.wantOutput {
				t.Errorf("stepTemplateRender() gotOutput = %v, want %v", gotOutput, tt.wantOutput)
			}

			if diff := cmp.Diff(gotOutput, tt.wantOutput); len(diff) != 0 {
				t.Errorf("%T differ (-got, +expected) %v", tt.wantOutput, diff)
			}
		})
	}
}

func Test_wrapWithCredential(t *testing.T) {
	type args struct {
		secretType string
		secretName string
		target     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "unknown secret type",
		args: args{
			secretType: "unknown",
			target:     "echo 1",
		},
		want: "echo 1",
	}, {
		name: "secret as kubeconfig type",
		args: args{
			secretType: string(SecretTypeKubeConfig),
			secretName: "config",
			target:     "echo 1",
		},
		want: readFile("testdata/credential-kubeconfig.json"),
	}, {
		name: "secret as secret text type",
		args: args{
			secretType: string(SecretTypeSecretText),
			secretName: "config",
			target:     "echo 1",
		},
		want: readFile("testdata/credential-string.json"),
	}, {
		name: "secret as ssh auth type",
		args: args{
			secretType: string(SecretTypeSSHAuth),
			secretName: "config",
			target:     "echo 1",
		},
		want: readFile("testdata/credential-ssh.json"),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wrapWithCredential(tt.args.secretType, tt.args.secretName, tt.args.target); got != tt.want {
				t.Errorf("wrapWithCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONFormat(t *testing.T) {
	assert.Equal(t, "abc", jsonFormat("abc"))
	assert.Equal(t, "abc", jsonFormat(" abc "))
}

func readFile(file string) string {
	if data, err := ioutil.ReadFile(file); err == nil {
		return string(data)
	}
	return ""
}
