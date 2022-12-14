/*
 *
 * Copyright 2019 The Tekton Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resources

import (
	"fmt"

	"github.com/ouyang-xlauncher/pipeline/pkg/apis/pipeline"
	"github.com/ouyang-xlauncher/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/ouyang-xlauncher/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/ouyang-xlauncher/pipeline/pkg/apis/resource"
	resourcev1alpha1 "github.com/ouyang-xlauncher/pipeline/pkg/apis/resource/v1alpha1"
	"github.com/ouyang-xlauncher/pipeline/pkg/names"
	corev1 "k8s.io/api/core/v1"
)

const (
	// unnamedCheckNamePrefix is the prefix added to the name of a condition's
	// spec.Check.Image if the name is missing
	unnamedCheckNamePrefix = "condition-check-"
)

// GetCondition is a function used to retrieve PipelineConditions.
type GetCondition func(string) (*v1alpha1.Condition, error)

// ResolvedConditionCheck contains a Condition and its associated ConditionCheck, if it
// exists. ConditionCheck can be nil to represent there being no ConditionCheck (i.e the condition
// has not been evaluated).
type ResolvedConditionCheck struct {
	ConditionRegisterName string
	PipelineTaskCondition *v1beta1.PipelineTaskCondition
	ConditionCheckName    string
	Condition             *v1alpha1.Condition
	ConditionCheck        *v1beta1.ConditionCheck
	// Resolved resources is a map of pipeline resources for this condition
	// keyed by the bound resource name (i.e. the name used in PipelineTaskCondition.Resources)
	ResolvedResources map[string]*resourcev1alpha1.PipelineResource

	images pipeline.Images
}

// TaskConditionCheckState is a slice of ResolvedConditionCheck the represents the current execution
// state of Conditions for a Task in a pipeline run.
type TaskConditionCheckState []*ResolvedConditionCheck

// HasStarted returns true if the conditionChecks for a given object have been created
func (state TaskConditionCheckState) HasStarted() bool {
	hasStarted := true
	for _, j := range state {
		if j.ConditionCheck == nil {
			hasStarted = false
		}
	}
	return hasStarted
}

// IsDone returns true if the status for all conditionChecks for a task indicate that they are done
func (state TaskConditionCheckState) IsDone() bool {
	if !state.HasStarted() {
		return false
	}
	isDone := true
	for _, rcc := range state {
		isDone = isDone && rcc.ConditionCheck.IsDone()
	}
	return isDone
}

// IsSuccess returns true if the status for all conditionChecks for a task indicate they have
// completed successfully
func (state TaskConditionCheckState) IsSuccess() bool {
	if !state.IsDone() {
		return false
	}
	isSuccess := true
	for _, rcc := range state {
		isSuccess = isSuccess && rcc.ConditionCheck.IsSuccessful()
	}
	return isSuccess
}

// ConditionToTaskSpec creates a TaskSpec from a given Condition
func (rcc *ResolvedConditionCheck) ConditionToTaskSpec() (*v1beta1.TaskSpec, error) {
	if rcc.Condition.Spec.Check.Name == "" {
		rcc.Condition.Spec.Check.Name = names.SimpleNameGenerator.RestrictLength(unnamedCheckNamePrefix + rcc.Condition.Name)
	}

	t := &v1beta1.TaskSpec{
		Steps:  []v1beta1.Step{rcc.Condition.Spec.Check},
		Params: rcc.Condition.Spec.Params,
	}

	for _, r := range rcc.Condition.Spec.Resources {
		if t.Resources == nil {
			t.Resources = &v1beta1.TaskResources{}
		}
		t.Resources.Inputs = append(t.Resources.Inputs, v1beta1.TaskResource{
			ResourceDeclaration: r,
		})
	}

	convertParamTemplates(&t.Steps[0], rcc.Condition.Spec.Params)
	err := applyResourceSubstitution(&t.Steps[0], rcc.ResolvedResources, rcc.Condition.Spec.Resources, rcc.images)

	if err != nil {
		return nil, fmt.Errorf("failed to replace resource template strings %w", err)
	}

	return t, nil
}

// convertParamTemplates replaces all instances of $(params.x) in the container to $(inputs.params.x) for each param name.
func convertParamTemplates(step *v1beta1.Step, params []v1beta1.ParamSpec) {
	replacements := make(map[string]string)
	for _, p := range params {
		replacements[fmt.Sprintf("params.%s", p.Name)] = fmt.Sprintf("$(inputs.params.%s)", p.Name)
		v1beta1.ApplyStepReplacements(step, replacements, map[string][]string{})
	}

	v1beta1.ApplyStepReplacements(step, replacements, map[string][]string{})
}

// applyResourceSubstitution applies the substitution from values in resources which are referenced
// in spec as subitems of the replacementStr.
func applyResourceSubstitution(step *v1beta1.Step, resolvedResources map[string]*resourcev1alpha1.PipelineResource, conditionResources []v1beta1.ResourceDeclaration, images pipeline.Images) error {
	replacements := make(map[string]string)
	for _, cr := range conditionResources {
		if rSpec, ok := resolvedResources[cr.Name]; ok {
			r, err := resource.FromType(cr.Name, rSpec, images)
			if err != nil {
				return fmt.Errorf("error trying to create resource: %w", err)
			}
			for k, v := range r.Replacements() {
				replacements[fmt.Sprintf("resources.%s.%s", cr.Name, k)] = v
			}
			replacements[fmt.Sprintf("resources.%s.path", cr.Name)] = v1beta1.InputResourcePath(cr)
		}
	}
	v1beta1.ApplyStepReplacements(step, replacements, map[string][]string{})
	return nil
}

// NewConditionCheckStatus creates a ConditionCheckStatus from a ConditionCheck
func (rcc *ResolvedConditionCheck) NewConditionCheckStatus() *v1beta1.ConditionCheckStatus {
	var checkStep corev1.ContainerState
	trs := rcc.ConditionCheck.Status
	for _, s := range trs.Steps {
		if s.Name == rcc.Condition.Spec.Check.Name {
			checkStep = s.ContainerState
			break
		}
	}

	return &v1beta1.ConditionCheckStatus{
		Status: trs.Status,
		ConditionCheckStatusFields: v1beta1.ConditionCheckStatusFields{
			PodName:        trs.PodName,
			StartTime:      trs.StartTime,
			CompletionTime: trs.CompletionTime,
			Check:          checkStep,
		},
	}
}

// ToTaskResourceBindings converts pipeline resources in a ResolvedConditionCheck to a list of TaskResourceBinding
// and returns them.
func (rcc *ResolvedConditionCheck) ToTaskResourceBindings() []v1beta1.TaskResourceBinding {
	var trb []v1beta1.TaskResourceBinding

	for name, r := range rcc.ResolvedResources {
		tr := v1beta1.TaskResourceBinding{
			PipelineResourceBinding: v1beta1.PipelineResourceBinding{
				Name: name,
			},
		}
		if r.SelfLink != "" {
			tr.ResourceRef = &v1beta1.PipelineResourceRef{
				Name:       r.Name,
				APIVersion: r.APIVersion,
			}
		} else if r.Spec.Type != "" {
			tr.ResourceSpec = &resourcev1alpha1.PipelineResourceSpec{
				Type:         r.Spec.Type,
				Params:       r.Spec.Params,
				SecretParams: r.Spec.SecretParams,
			}
		}
		trb = append(trb, tr)
	}

	return trb
}
