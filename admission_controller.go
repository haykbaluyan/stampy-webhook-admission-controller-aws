package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"git.soma.salesforce.com/stampy-webhook-admission-controller-aws/validator"
	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// AdmissionControllerInterface exposes admission controller related operations
type AdmissionControllerInterface interface {
	Mutate(ar *v1beta1.AdmissionReview) (r *v1beta1.AdmissionResponse)
}

// AdmissionController implements admission controller related operations for AWS
type admissionController struct {
	logger *logrus.Logger
	region string // aws region that stores signatures
	bucket string // aws s3 bucket that stores signatures
}

// NewAdmissionController constructor
func NewAdmissionController(region, bucket string, logger *logrus.Logger) (AdmissionControllerInterface, error) {
	ac := new(admissionController)
	ac.region = region
	ac.bucket = bucket
	ac.logger = logger
	return ac, nil
}

// Mutate implements mutating webhook
func (ac *admissionController) Mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	var deployment corev1.Deployment
	if err := json.Unmarshal(ar.Request.Object.Raw, &deployment); err != nil {
		ac.logger.Errorf("api=mutate, reason='could not unmarshal raw object: %v'", err)
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: fmt.Sprintf("could not decode admission request object"),
			},
		}
	}

	imageManager := NewImageController(ac.region, ac.bucket, ac.logger)
	var patch []patchOperation
	containers := deployment.Spec.Template.Spec.Containers
	for i, container := range containers {
		image := container.Image
		host, repo, tag := parseImage(image)
		manifest, err := imageManager.GetManifest(repo, tag)
		if err != nil {
			ac.logger.Errorf("api=mutate, reason=GetManifest, repo=%q, tag=%q, err=%v", repo, tag, err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: fmt.Sprintf("failed to fetch manifest, repo=%q, tag=%q", repo, tag),
				},
			}
		}
		ac.logger.Infof("manifest %q", manifest)

		manifestDigest := validator.SHA256Digest([]byte(manifest))
		manifestSig, err := imageManager.GetManifestSignature(repo, manifestDigest)
		if err != nil {
			ac.logger.Errorf("api=mutate, reason=GetManifestSignature, repo=%q, tag=%q, err=%v", repo, tag, err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: fmt.Sprintf("failed to fetch manifest signature, repo=%q, tag=%q", repo, tag),
				},
			}
		}

		if len(manifestSig) == 0 {
			ac.logger.Errorf("api=mutate, reason='empty manifest signature', repo=%q, tag=%q, manifest_digest=%q, err=%v", repo, tag, manifestDigest, err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: fmt.Sprintf("failed to fetch manifest signature, repo=%q, tag=%q", repo, tag),
				},
			}
		}

		status, manifestDigest, err := validator.ValidateManifestSignature(manifest, manifestSig)
		if err != nil {
			ac.logger.Errorf("api=mutate, reason=ValidateManifestSignature, repo=%q, tag=%q, err=%v", repo, tag, err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: fmt.Sprintf("failed to validate manifest signature, repo=%q, tag=%q", repo, tag),
				},
			}
		}

		if !status {
			ac.logger.Errorf("api=mutate, reason=ValidateManifestSignature, status=%t, repo=%q, tag=%q, err=%v", status, repo, tag, err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: fmt.Sprintf("failed to validate manifest signature, repo=%q, tag=%q", repo, tag),
				},
			}
		}

		patch = append(patch, patchOperation{
			Op:    "replace",
			Path:  fmt.Sprintf("/spec/template/spec/containers/%d/image", i),
			Value: fmt.Sprintf("%s/%s@sha256:%s", host, repo, manifestDigest),
		})
	}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	ac.logger.Infof("api=mutate, admissionResponse_patch=%v\n", string(patchBytes))
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func parseImage(image string) (host, repo, tag string) {
	s := strings.SplitN(image, "/", 2)
	host = s[0]

	imagePath := ""
	if len(s) > 1 {
		imagePath = s[1]
	}

	s = strings.SplitN(imagePath, "@", 2)
	repo = s[0]
	if len(s) > 1 {
		tag = s[1]
		return host, repo, tag
	}

	s = strings.SplitN(imagePath, ":", 2)
	repo = s[0]
	if len(s) > 1 {
		tag = s[1]
	} else {
		tag = "latest"
	}
	return host, repo, tag
}
