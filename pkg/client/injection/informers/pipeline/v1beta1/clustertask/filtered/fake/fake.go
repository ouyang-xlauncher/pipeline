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

// Code generated by injection-gen. DO NOT EDIT.

package fake

import (
	context "context"

	factoryfiltered "github.com/ouyang-xlauncher/pipeline/pkg/client/injection/informers/factory/filtered"
	filtered "github.com/ouyang-xlauncher/pipeline/pkg/client/injection/informers/pipeline/v1beta1/clustertask/filtered"
	controller "knative.dev/pkg/controller"
	injection "knative.dev/pkg/injection"
	logging "knative.dev/pkg/logging"
)

var Get = filtered.Get

func init() {
	injection.Fake.RegisterFilteredInformers(withInformer)
}

func withInformer(ctx context.Context) (context.Context, []controller.Informer) {
	untyped := ctx.Value(factoryfiltered.LabelKey{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch labelkey from context.")
	}
	labelSelectors := untyped.([]string)
	infs := []controller.Informer{}
	for _, selector := range labelSelectors {
		f := factoryfiltered.Get(ctx, selector)
		inf := f.Tekton().V1beta1().ClusterTasks()
		ctx = context.WithValue(ctx, filtered.Key{Selector: selector}, inf)
		infs = append(infs, inf.Informer())
	}
	return ctx, infs
}
