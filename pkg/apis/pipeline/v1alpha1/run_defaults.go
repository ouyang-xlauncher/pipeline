/*
Copyright 2020 The Tekton Authors

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

package v1alpha1

import (
	"context"

	"github.com/ouyang-xlauncher/pipeline/pkg/apis/config"
	"knative.dev/pkg/apis"
)

var _ apis.Defaultable = (*Run)(nil)

func (r *Run) SetDefaults(ctx context.Context) {
	ctx = apis.WithinParent(ctx, r.ObjectMeta)
	r.Spec.SetDefaults(apis.WithinSpec(ctx))
}

func (rs *RunSpec) SetDefaults(ctx context.Context) {
	cfg := config.FromContextOrDefaults(ctx)
	defaultSA := cfg.Defaults.DefaultServiceAccount
	if rs.ServiceAccountName == "" && defaultSA != "" {
		rs.ServiceAccountName = defaultSA
	}
	defaultPodTemplate := cfg.Defaults.DefaultPodTemplate
	if rs.PodTemplate == nil {
		rs.PodTemplate = defaultPodTemplate
	}
}
