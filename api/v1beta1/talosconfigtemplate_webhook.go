// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package v1beta1

import (
	"context"

	"github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/cluster-api/util/topology"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (r *TalosConfigTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:verbs=create;update,path=/validate-bootstrap-cluster-x-k8s-io-v1beta1-talosconfigtemplate,mutating=false,failurePolicy=fail,groups=bootstrap.cluster.x-k8s.io,resources=talosconfigtemplates,versions=v1beta1,name=vtalosconfigtemplate.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1

var _ admission.Validator[*TalosConfigTemplate] = &TalosConfigTemplate{}

// ValidateCreate implements admission.Validator so a webhook will be registered for the type
func (r *TalosConfigTemplate) ValidateCreate(ctx context.Context, obj *TalosConfigTemplate) (admission.Warnings, error) {
	return nil, validateTemplate(obj, nil)
}

// ValidateUpdate implements admission.Validator so a webhook will be registered for the type
func (r *TalosConfigTemplate) ValidateUpdate(ctx context.Context, oldObj *TalosConfigTemplate, newObj *TalosConfigTemplate) (admission.Warnings, error) {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var allErrs field.ErrorList

	// Every TalosConfig generated from the template carries its spec, so the spec is frozen. The
	// template metadata only reaches configs generated later, so it stays mutable, as it does on
	// Cluster API's KubeadmConfigTemplate. The topology controller's dry-run requests are exempt (#257).
	if !topology.IsDryRunRequest(req, newObj) && !cmp.Equal(newObj.Spec.Template.Spec, oldObj.Spec.Template.Spec) {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec", "template", "spec"), "TalosConfigTemplate spec.template.spec is immutable"))
	}

	return nil, validateTemplate(newObj, allErrs)
}

// ValidateDelete implements admission.Validator so a webhook will be registered for the type
func (r *TalosConfigTemplate) ValidateDelete(ctx context.Context, obj *TalosConfigTemplate) (admission.Warnings, error) {
	return nil, nil
}

func validateTemplate(t *TalosConfigTemplate, allErrs field.ErrorList) error {
	allErrs = append(allErrs, t.Spec.Template.ObjectMeta.Validate(field.NewPath("spec", "template", "metadata"))...)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind("TalosConfigTemplate").GroupKind(), t.Name, allErrs)
}
