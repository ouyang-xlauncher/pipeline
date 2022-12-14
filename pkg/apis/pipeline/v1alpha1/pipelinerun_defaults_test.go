/*
Copyright 2019 The Tekton Authors

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

package v1alpha1_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ouyang-xlauncher/pipeline/pkg/apis/config"
	"github.com/ouyang-xlauncher/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/ouyang-xlauncher/pipeline/test/diff"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logtesting "knative.dev/pkg/logging/testing"
)

var (
	ignoreUnexportedResources = cmpopts.IgnoreUnexported()
)

func TestPipelineRunSpec_SetDefaults(t *testing.T) {
	cases := []struct {
		desc string
		prs  *v1alpha1.PipelineRunSpec
		want *v1alpha1.PipelineRunSpec
	}{
		{
			desc: "timeout is nil",
			prs:  &v1alpha1.PipelineRunSpec{},
			want: &v1alpha1.PipelineRunSpec{
				ServiceAccountName: config.DefaultServiceAccountValue,
				Timeout:            &metav1.Duration{Duration: config.DefaultTimeoutMinutes * time.Minute},
			},
		},
		{
			desc: "timeout is not nil",
			prs: &v1alpha1.PipelineRunSpec{
				Timeout: &metav1.Duration{Duration: 500 * time.Millisecond},
			},
			want: &v1alpha1.PipelineRunSpec{
				ServiceAccountName: config.DefaultServiceAccountValue,
				Timeout:            &metav1.Duration{Duration: 500 * time.Millisecond},
			},
		},
		{
			desc: "pod template is nil",
			prs:  &v1alpha1.PipelineRunSpec{},
			want: &v1alpha1.PipelineRunSpec{
				ServiceAccountName: config.DefaultServiceAccountValue,
				Timeout:            &metav1.Duration{Duration: config.DefaultTimeoutMinutes * time.Minute},
			},
		},
		{
			desc: "pod template is not nil",
			prs: &v1alpha1.PipelineRunSpec{
				PodTemplate: &v1alpha1.PodTemplate{
					NodeSelector: map[string]string{
						"label": "value",
					},
				},
			},
			want: &v1alpha1.PipelineRunSpec{
				ServiceAccountName: config.DefaultServiceAccountValue,
				Timeout:            &metav1.Duration{Duration: config.DefaultTimeoutMinutes * time.Minute},
				PodTemplate: &v1alpha1.PodTemplate{
					NodeSelector: map[string]string{
						"label": "value",
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			tc.prs.SetDefaults(ctx)

			if d := cmp.Diff(tc.want, tc.prs); d != "" {
				t.Errorf("Mismatch of PipelineRunSpec %s", diff.PrintWantGot(d))
			}
		})
	}

}

func TestPipelineRunDefaulting(t *testing.T) {
	tests := []struct {
		name string
		in   *v1alpha1.PipelineRun
		want *v1alpha1.PipelineRun
		wc   func(context.Context) context.Context
	}{{
		name: "empty no context",
		in:   &v1alpha1.PipelineRun{},
		want: &v1alpha1.PipelineRun{
			Spec: v1alpha1.PipelineRunSpec{
				ServiceAccountName: config.DefaultServiceAccountValue,
				Timeout:            &metav1.Duration{Duration: config.DefaultTimeoutMinutes * time.Minute},
			},
		},
	}, {
		name: "PipelineRef default config context",
		in: &v1alpha1.PipelineRun{
			Spec: v1alpha1.PipelineRunSpec{
				PipelineRef: &v1alpha1.PipelineRef{Name: "foo"},
			},
		},
		want: &v1alpha1.PipelineRun{
			Spec: v1alpha1.PipelineRunSpec{
				PipelineRef:        &v1alpha1.PipelineRef{Name: "foo"},
				ServiceAccountName: config.DefaultServiceAccountValue,
				Timeout:            &metav1.Duration{Duration: 5 * time.Minute},
			},
		},
		wc: func(ctx context.Context) context.Context {
			s := config.NewStore(logtesting.TestLogger(t))
			s.OnConfigChanged(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: config.GetDefaultsConfigName(),
				},
				Data: map[string]string{
					"default-timeout-minutes": "5",
				},
			})
			return s.ToContext(ctx)
		},
	}, {
		name: "PipelineRef default config context with sa",
		in: &v1alpha1.PipelineRun{
			Spec: v1alpha1.PipelineRunSpec{
				PipelineRef: &v1alpha1.PipelineRef{Name: "foo"},
			},
		},
		want: &v1alpha1.PipelineRun{
			Spec: v1alpha1.PipelineRunSpec{
				PipelineRef:        &v1alpha1.PipelineRef{Name: "foo"},
				Timeout:            &metav1.Duration{Duration: 5 * time.Minute},
				ServiceAccountName: "tekton",
			},
		},
		wc: func(ctx context.Context) context.Context {
			s := config.NewStore(logtesting.TestLogger(t))
			s.OnConfigChanged(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: config.GetDefaultsConfigName(),
				},
				Data: map[string]string{
					"default-timeout-minutes": "5",
					"default-service-account": "tekton",
				},
			})
			return s.ToContext(ctx)
		},
	}, {
		name: "PipelineRef pod template is coming from default config pod template",
		in: &v1alpha1.PipelineRun{
			Spec: v1alpha1.PipelineRunSpec{
				PipelineRef: &v1alpha1.PipelineRef{Name: "foo"},
			},
		},
		want: &v1alpha1.PipelineRun{
			Spec: v1alpha1.PipelineRunSpec{
				PipelineRef:        &v1alpha1.PipelineRef{Name: "foo"},
				Timeout:            &metav1.Duration{Duration: 5 * time.Minute},
				ServiceAccountName: "tekton",
				PodTemplate: &v1alpha1.PodTemplate{
					NodeSelector: map[string]string{
						"label": "value",
					},
				},
			},
		},
		wc: func(ctx context.Context) context.Context {
			s := config.NewStore(logtesting.TestLogger(t))
			s.OnConfigChanged(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: config.GetDefaultsConfigName(),
				},
				Data: map[string]string{
					"default-timeout-minutes": "5",
					"default-service-account": "tekton",
					"default-pod-template":    "nodeSelector: { 'label': 'value' }",
				},
			})
			return s.ToContext(ctx)
		},
	}, {
		name: "PipelineRef pod template takes precedence over default config pod template",
		in: &v1alpha1.PipelineRun{
			Spec: v1alpha1.PipelineRunSpec{
				PipelineRef: &v1alpha1.PipelineRef{Name: "foo"},
				PodTemplate: &v1alpha1.PodTemplate{
					NodeSelector: map[string]string{
						"label2": "value2",
					},
				},
			},
		},
		want: &v1alpha1.PipelineRun{
			Spec: v1alpha1.PipelineRunSpec{
				PipelineRef:        &v1alpha1.PipelineRef{Name: "foo"},
				Timeout:            &metav1.Duration{Duration: 5 * time.Minute},
				ServiceAccountName: "tekton",
				PodTemplate: &v1alpha1.PodTemplate{
					NodeSelector: map[string]string{
						"label2": "value2",
					},
				},
			},
		},
		wc: func(ctx context.Context) context.Context {
			s := config.NewStore(logtesting.TestLogger(t))
			s.OnConfigChanged(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: config.GetDefaultsConfigName(),
				},
				Data: map[string]string{
					"default-timeout-minutes": "5",
					"default-service-account": "tekton",
					"default-pod-template":    "nodeSelector: { 'label': 'value' }",
				},
			})
			return s.ToContext(ctx)
		},
	}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in
			ctx := context.Background()
			if tc.wc != nil {
				ctx = tc.wc(ctx)
			}
			got.SetDefaults(ctx)
			if !cmp.Equal(got, tc.want, ignoreUnexportedResources) {
				d := cmp.Diff(got, tc.want, ignoreUnexportedResources)
				t.Errorf("SetDefaults %s", diff.PrintWantGot(d))
			}
		})
	}
}
