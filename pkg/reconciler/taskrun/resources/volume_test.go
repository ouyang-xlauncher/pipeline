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

package resources_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ouyang-xlauncher/pipeline/pkg/reconciler/taskrun/resources"
	"github.com/ouyang-xlauncher/pipeline/test/diff"
	corev1 "k8s.io/api/core/v1"
)

func TestGetPVCVolume(t *testing.T) {
	expectedVolume := corev1.Volume{
		Name: "test-pvc",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "test-pvc"},
		},
	}
	if d := cmp.Diff(expectedVolume, resources.GetPVCVolume("test-pvc")); d != "" {
		t.Fatalf("PVC volume mismatch: %s", diff.PrintWantGot(d))
	}
}
