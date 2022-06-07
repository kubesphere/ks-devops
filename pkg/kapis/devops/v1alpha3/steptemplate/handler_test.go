package steptemplate

import (
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/devops/pkg/api/devops/v1alpha3"
	"testing"
)

func Test_handler_stepTemplateRender(t *testing.T) {
	stepTemplate := &v1alpha3.StepTemplateSpec{
		Template: `echo 1`,
		Runtime:  "shell",
	}
	stepTemplateWithParameters := &v1alpha3.StepTemplateSpec{
		Template: `docker login -u $USERNAMEVARIABLE -p $PASSWORDVARIABLE
docker build {{.param.context}} -t {{.param.tag}}`,
		Runtime:   "shell",
		Container: "base",
		Secret: v1alpha3.SecretInStep{
			Wrap: true,
			Type: string(v1.SecretTypeBasicAuth),
		},
		Parameters: []v1alpha3.ParameterInStep{{
			Name: "context",
		}, {
			Name: "tag",
		}},
	}
	stepTemplateWithDSL := &v1alpha3.StepTemplateSpec{
		Template: `echo "1"`,
		Runtime:  "dsl",
	}

	type args struct {
		stepTemplate *v1alpha3.StepTemplateSpec
		param        map[string]string
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
		wantOutput: `sh '''
echo 1
'''`,
		wantErr: false,
	}, {
		name: "docker build command with parameters",
		args: args{
			stepTemplate: stepTemplateWithParameters,
			param: map[string]string{
				"context": "dir",
				"tag":     "image:tag",
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
		wantOutput: `container("base") {
	withCredential[usernamePassword(credentialsId : "docker" ,passwordVariable : 'PASSWORDVARIABLE' ,usernameVariable : 'USERNAMEVARIABLE')]) {
	sh '''
docker login -u $USERNAMEVARIABLE -p $PASSWORDVARIABLE
docker build dir -t image:tag
'''
}
}`,
		wantErr: false,
	}, {
		name: "a simple dsl without any parameters",
		args: args{
			stepTemplate: stepTemplateWithDSL,
		},
		wantOutput: `echo "1"`,
		wantErr:    false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutput, err := stepTemplateRender(tt.args.stepTemplate, tt.args.param, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("stepTemplateRender() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOutput != tt.wantOutput {
				t.Errorf("stepTemplateRender() gotOutput = %v, want %v", gotOutput, tt.wantOutput)
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
			secretType: string(v1alpha3.SecretTypeKubeConfig),
			secretName: "config",
			target:     "echo 1",
		},
		want: `withCredential[kubeconfigContent(credentialsId : "config" ,variable : 'VARIABLE')]) {
	echo 1
}`,
	}, {
		name: "secret as secret text type",
		args: args{
			secretType: string(v1alpha3.SecretTypeSecretText),
			secretName: "config",
			target:     "echo 1",
		},
		want: `withCredential[string(credentialsId : "config" ,variable : 'VARIABLE')]) {
	echo 1
}`,
	}, {
		name: "secret as ssh auth type",
		args: args{
			secretType: string(v1alpha3.SecretTypeSSHAuth),
			secretName: "config",
			target:     "echo 1",
		},
		want: `withCredential[sshUserPrivateKey(credentialsId : "config" ,keyFileVariable : 'KEYFILEVARIABLE' ,passphraseVariable : 'PASSPHRASEVARIABLE' ,usernameVariable : 'SSHUSERPRIVATEKEY')]) {
	echo 1
}`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wrapWithCredential(tt.args.secretType, tt.args.secretName, tt.args.target); got != tt.want {
				t.Errorf("wrapWithCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}
