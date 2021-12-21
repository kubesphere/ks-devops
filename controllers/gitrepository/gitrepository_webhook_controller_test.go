package gitrepository

import (
	"github.com/drone/go-scm/scm"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"testing"
)

func Test_getRepo(t *testing.T) {
	type args struct {
		repo *v1alpha1.GitRepository
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "not supported provider",
		args: args{
			repo: &v1alpha1.GitRepository{Spec: v1alpha1.GitRepositorySpec{Provider: "fake"}},
		},
		want: "",
	}, {
		name: "provider is emtpy",
		args: args{
			repo: &v1alpha1.GitRepository{Spec: v1alpha1.GitRepositorySpec{
				URL: "https://github.com/linuxsuren/test",
			}},
		},
		want: "",
	}, {
		name: "github as the provider",
		args: args{
			repo: &v1alpha1.GitRepository{Spec: v1alpha1.GitRepositorySpec{
				Provider: "github",
				URL:      "https://github.com/linuxsuren/test",
			}},
		},
		want: "linuxsuren/test",
	}, {
		name: "gitlab as the provider",
		args: args{
			repo: &v1alpha1.GitRepository{Spec: v1alpha1.GitRepositorySpec{
				Provider: "gitlab",
				URL:      "https://gitlab.com/linuxsuren/test",
			}},
		},
		want: "linuxsuren/test",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRepo(tt.args.repo); got != tt.want {
				t.Errorf("getRepo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_exist(t *testing.T) {
	type args struct {
		server string
		hooks  []*scm.Hook
	}
	tests := []struct {
		name      string
		args      args
		wantExist bool
		wantId    string
	}{{
		name: "not exist from empty",
		args: args{
			server: "fake",
			hooks:  nil,
		},
		wantExist: false,
		wantId:    "",
	}, {
		name: "not exist",
		args: args{
			server: "fake",
			hooks: []*scm.Hook{{
				Target: "random",
			}},
		},
		wantExist: false,
		wantId:    "",
	}, {
		name: "exist",
		args: args{
			server: "fake",
			hooks: []*scm.Hook{{
				ID:     "fake-id",
				Target: "fake",
			}, {
				ID:     "random-id",
				Target: "random",
			}},
		},
		wantExist: true,
		wantId:    "fake-id",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExist, gotId := exist(tt.args.server, tt.args.hooks)
			if gotExist != tt.wantExist {
				t.Errorf("exist() gotExist = %v, want %v", gotExist, tt.wantExist)
			}
			if gotId != tt.wantId {
				t.Errorf("exist() gotId = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}
