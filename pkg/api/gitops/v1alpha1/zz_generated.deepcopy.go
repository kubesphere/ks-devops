//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Application) DeepCopyInto(out *Application) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Application.
func (in *Application) DeepCopy() *Application {
	if in == nil {
		return nil
	}
	out := new(Application)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Application) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationDestination) DeepCopyInto(out *ApplicationDestination) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationDestination.
func (in *ApplicationDestination) DeepCopy() *ApplicationDestination {
	if in == nil {
		return nil
	}
	out := new(ApplicationDestination)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationList) DeepCopyInto(out *ApplicationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Application, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationList.
func (in *ApplicationList) DeepCopy() *ApplicationList {
	if in == nil {
		return nil
	}
	out := new(ApplicationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApplicationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationSource) DeepCopyInto(out *ApplicationSource) {
	*out = *in
	if in.Helm != nil {
		in, out := &in.Helm, &out.Helm
		*out = new(ApplicationSourceHelm)
		(*in).DeepCopyInto(*out)
	}
	if in.Kustomize != nil {
		in, out := &in.Kustomize, &out.Kustomize
		*out = new(ApplicationSourceKustomize)
		(*in).DeepCopyInto(*out)
	}
	if in.Ksonnet != nil {
		in, out := &in.Ksonnet, &out.Ksonnet
		*out = new(ApplicationSourceKsonnet)
		(*in).DeepCopyInto(*out)
	}
	if in.Directory != nil {
		in, out := &in.Directory, &out.Directory
		*out = new(ApplicationSourceDirectory)
		(*in).DeepCopyInto(*out)
	}
	if in.Plugin != nil {
		in, out := &in.Plugin, &out.Plugin
		*out = new(ApplicationSourcePlugin)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationSource.
func (in *ApplicationSource) DeepCopy() *ApplicationSource {
	if in == nil {
		return nil
	}
	out := new(ApplicationSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationSourceDirectory) DeepCopyInto(out *ApplicationSourceDirectory) {
	*out = *in
	in.Jsonnet.DeepCopyInto(&out.Jsonnet)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationSourceDirectory.
func (in *ApplicationSourceDirectory) DeepCopy() *ApplicationSourceDirectory {
	if in == nil {
		return nil
	}
	out := new(ApplicationSourceDirectory)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationSourceHelm) DeepCopyInto(out *ApplicationSourceHelm) {
	*out = *in
	if in.ValueFiles != nil {
		in, out := &in.ValueFiles, &out.ValueFiles
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make([]HelmParameter, len(*in))
		copy(*out, *in)
	}
	if in.FileParameters != nil {
		in, out := &in.FileParameters, &out.FileParameters
		*out = make([]HelmFileParameter, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationSourceHelm.
func (in *ApplicationSourceHelm) DeepCopy() *ApplicationSourceHelm {
	if in == nil {
		return nil
	}
	out := new(ApplicationSourceHelm)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationSourceJsonnet) DeepCopyInto(out *ApplicationSourceJsonnet) {
	*out = *in
	if in.ExtVars != nil {
		in, out := &in.ExtVars, &out.ExtVars
		*out = make([]JsonnetVar, len(*in))
		copy(*out, *in)
	}
	if in.TLAs != nil {
		in, out := &in.TLAs, &out.TLAs
		*out = make([]JsonnetVar, len(*in))
		copy(*out, *in)
	}
	if in.Libs != nil {
		in, out := &in.Libs, &out.Libs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationSourceJsonnet.
func (in *ApplicationSourceJsonnet) DeepCopy() *ApplicationSourceJsonnet {
	if in == nil {
		return nil
	}
	out := new(ApplicationSourceJsonnet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationSourceKsonnet) DeepCopyInto(out *ApplicationSourceKsonnet) {
	*out = *in
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make([]KsonnetParameter, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationSourceKsonnet.
func (in *ApplicationSourceKsonnet) DeepCopy() *ApplicationSourceKsonnet {
	if in == nil {
		return nil
	}
	out := new(ApplicationSourceKsonnet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationSourceKustomize) DeepCopyInto(out *ApplicationSourceKustomize) {
	*out = *in
	if in.Images != nil {
		in, out := &in.Images, &out.Images
		*out = make(KustomizeImages, len(*in))
		copy(*out, *in)
	}
	if in.CommonLabels != nil {
		in, out := &in.CommonLabels, &out.CommonLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.CommonAnnotations != nil {
		in, out := &in.CommonAnnotations, &out.CommonAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationSourceKustomize.
func (in *ApplicationSourceKustomize) DeepCopy() *ApplicationSourceKustomize {
	if in == nil {
		return nil
	}
	out := new(ApplicationSourceKustomize)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationSourcePlugin) DeepCopyInto(out *ApplicationSourcePlugin) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(Env, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(EnvEntry)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationSourcePlugin.
func (in *ApplicationSourcePlugin) DeepCopy() *ApplicationSourcePlugin {
	if in == nil {
		return nil
	}
	out := new(ApplicationSourcePlugin)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationSpec) DeepCopyInto(out *ApplicationSpec) {
	*out = *in
	if in.ArgoApp != nil {
		in, out := &in.ArgoApp, &out.ArgoApp
		*out = new(ArgoApplication)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationSpec.
func (in *ApplicationSpec) DeepCopy() *ApplicationSpec {
	if in == nil {
		return nil
	}
	out := new(ApplicationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoApplication) DeepCopyInto(out *ArgoApplication) {
	*out = *in
	in.Source.DeepCopyInto(&out.Source)
	out.Destination = in.Destination
	if in.SyncPolicy != nil {
		in, out := &in.SyncPolicy, &out.SyncPolicy
		*out = new(SyncPolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.IgnoreDifferences != nil {
		in, out := &in.IgnoreDifferences, &out.IgnoreDifferences
		*out = make([]ResourceIgnoreDifferences, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Info != nil {
		in, out := &in.Info, &out.Info
		*out = make([]Info, len(*in))
		copy(*out, *in)
	}
	if in.RevisionHistoryLimit != nil {
		in, out := &in.RevisionHistoryLimit, &out.RevisionHistoryLimit
		*out = new(int64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoApplication.
func (in *ArgoApplication) DeepCopy() *ArgoApplication {
	if in == nil {
		return nil
	}
	out := new(ArgoApplication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Backoff) DeepCopyInto(out *Backoff) {
	*out = *in
	if in.Factor != nil {
		in, out := &in.Factor, &out.Factor
		*out = new(int64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Backoff.
func (in *Backoff) DeepCopy() *Backoff {
	if in == nil {
		return nil
	}
	out := new(Backoff)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Env) DeepCopyInto(out *Env) {
	{
		in := &in
		*out = make(Env, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(EnvEntry)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Env.
func (in Env) DeepCopy() Env {
	if in == nil {
		return nil
	}
	out := new(Env)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvEntry) DeepCopyInto(out *EnvEntry) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvEntry.
func (in *EnvEntry) DeepCopy() *EnvEntry {
	if in == nil {
		return nil
	}
	out := new(EnvEntry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmFileParameter) DeepCopyInto(out *HelmFileParameter) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmFileParameter.
func (in *HelmFileParameter) DeepCopy() *HelmFileParameter {
	if in == nil {
		return nil
	}
	out := new(HelmFileParameter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmParameter) DeepCopyInto(out *HelmParameter) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmParameter.
func (in *HelmParameter) DeepCopy() *HelmParameter {
	if in == nil {
		return nil
	}
	out := new(HelmParameter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Info) DeepCopyInto(out *Info) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Info.
func (in *Info) DeepCopy() *Info {
	if in == nil {
		return nil
	}
	out := new(Info)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JsonnetVar) DeepCopyInto(out *JsonnetVar) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JsonnetVar.
func (in *JsonnetVar) DeepCopy() *JsonnetVar {
	if in == nil {
		return nil
	}
	out := new(JsonnetVar)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KsonnetParameter) DeepCopyInto(out *KsonnetParameter) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KsonnetParameter.
func (in *KsonnetParameter) DeepCopy() *KsonnetParameter {
	if in == nil {
		return nil
	}
	out := new(KsonnetParameter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in KustomizeImages) DeepCopyInto(out *KustomizeImages) {
	{
		in := &in
		*out = make(KustomizeImages, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KustomizeImages.
func (in KustomizeImages) DeepCopy() KustomizeImages {
	if in == nil {
		return nil
	}
	out := new(KustomizeImages)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceIgnoreDifferences) DeepCopyInto(out *ResourceIgnoreDifferences) {
	*out = *in
	if in.JSONPointers != nil {
		in, out := &in.JSONPointers, &out.JSONPointers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.JQPathExpressions != nil {
		in, out := &in.JQPathExpressions, &out.JQPathExpressions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ManagedFieldsManagers != nil {
		in, out := &in.ManagedFieldsManagers, &out.ManagedFieldsManagers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceIgnoreDifferences.
func (in *ResourceIgnoreDifferences) DeepCopy() *ResourceIgnoreDifferences {
	if in == nil {
		return nil
	}
	out := new(ResourceIgnoreDifferences)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RetryStrategy) DeepCopyInto(out *RetryStrategy) {
	*out = *in
	if in.Backoff != nil {
		in, out := &in.Backoff, &out.Backoff
		*out = new(Backoff)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RetryStrategy.
func (in *RetryStrategy) DeepCopy() *RetryStrategy {
	if in == nil {
		return nil
	}
	out := new(RetryStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in SyncOptions) DeepCopyInto(out *SyncOptions) {
	{
		in := &in
		*out = make(SyncOptions, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SyncOptions.
func (in SyncOptions) DeepCopy() SyncOptions {
	if in == nil {
		return nil
	}
	out := new(SyncOptions)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SyncPolicy) DeepCopyInto(out *SyncPolicy) {
	*out = *in
	if in.Automated != nil {
		in, out := &in.Automated, &out.Automated
		*out = new(SyncPolicyAutomated)
		**out = **in
	}
	if in.SyncOptions != nil {
		in, out := &in.SyncOptions, &out.SyncOptions
		*out = make(SyncOptions, len(*in))
		copy(*out, *in)
	}
	if in.Retry != nil {
		in, out := &in.Retry, &out.Retry
		*out = new(RetryStrategy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SyncPolicy.
func (in *SyncPolicy) DeepCopy() *SyncPolicy {
	if in == nil {
		return nil
	}
	out := new(SyncPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SyncPolicyAutomated) DeepCopyInto(out *SyncPolicyAutomated) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SyncPolicyAutomated.
func (in *SyncPolicyAutomated) DeepCopy() *SyncPolicyAutomated {
	if in == nil {
		return nil
	}
	out := new(SyncPolicyAutomated)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SyncStrategy) DeepCopyInto(out *SyncStrategy) {
	*out = *in
	if in.Apply != nil {
		in, out := &in.Apply, &out.Apply
		*out = new(SyncStrategyApply)
		**out = **in
	}
	if in.Hook != nil {
		in, out := &in.Hook, &out.Hook
		*out = new(SyncStrategyHook)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SyncStrategy.
func (in *SyncStrategy) DeepCopy() *SyncStrategy {
	if in == nil {
		return nil
	}
	out := new(SyncStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SyncStrategyApply) DeepCopyInto(out *SyncStrategyApply) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SyncStrategyApply.
func (in *SyncStrategyApply) DeepCopy() *SyncStrategyApply {
	if in == nil {
		return nil
	}
	out := new(SyncStrategyApply)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SyncStrategyHook) DeepCopyInto(out *SyncStrategyHook) {
	*out = *in
	out.SyncStrategyApply = in.SyncStrategyApply
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SyncStrategyHook.
func (in *SyncStrategyHook) DeepCopy() *SyncStrategyHook {
	if in == nil {
		return nil
	}
	out := new(SyncStrategyHook)
	in.DeepCopyInto(out)
	return out
}
