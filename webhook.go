package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

type webhook struct {
	server *http.Server
}

// WhSvrParameters server parameters
type WhSvrParameters struct {
	port           int    // webhook server port
	certFile       string // path to the x509 certificate for https
	keyFile        string // path to the x509 private key matching `CertFile`
	sidecarCfgFile string // path to sidecar injector configuration file
}

// main mutation process
func (vh *webhook) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	glog.Info("api=mutate, reason=started")
	status := "Success"
	allowed := true
	glog.Infof("api=mutate, uid=%q, kind=%q, operation=%q, namespace=%q", ar.Request.UID, ar.Request.Kind, ar.Request.Operation, ar.Request.Namespace)
	if ar.Request.Kind.Kind != "Deployment" {
		admissionResp := &admissionv1beta1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: allowed,
			Result: &metav1.Status{
				Status:  status,
				Message: "ok",
			},
		}

		return admissionResp
	}

	var deployment appsv1.Deployment
	if err := json.Unmarshal(ar.Request.Object.Raw, &deployment); err != nil {
		glog.Errorf("api=mutate, reason='could not unmarshal raw object: %v'", err)
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: fmt.Sprintf("could not decode admission request object"),
			},
		}
	}

	var patch []patchOperation
	containers := deployment.Spec.Template.Spec.Containers
	for i, container := range containers {
		image := container.Image
		glog.Infof("api=validate, image=%q", image)

		patch = append(patch, patchOperation{
			Op:    "replace",
			Path:  fmt.Sprintf("/spec/template/spec/containers/%d/image", i),
			Value: image + ":trusty",
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

	glog.Infof("api=mutate, admissionResponse_patch=%v\n", string(patchBytes))
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// main mutation process
func (vh *webhook) validate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	glog.Info("api=validate, reason=started")
	status := "Success"
	allowed := true
	glog.Infof("api=validate, uid=%q, kind=%q, operation=%q, namespace=%q", ar.Request.UID, ar.Request.Kind, ar.Request.Operation, ar.Request.Namespace)
	if ar.Request.Kind.Kind != "Deployment" {
		admissionResp := &admissionv1beta1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: allowed,
			Result: &metav1.Status{
				Status:  status,
				Message: "ok",
			},
		}

		return admissionResp
	}

	var deployment appsv1.Deployment
	if err := json.Unmarshal(ar.Request.Object.Raw, &deployment); err != nil {
		glog.Errorf("api=validate, reason='Could not unmarshal raw object: %v'", err)
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: fmt.Sprintf("could not decode admission request object"),
			},
		}
	}

	containers := deployment.Spec.Template.Spec.Containers
	for _, container := range containers {
		image := container.Image
		glog.Infof("api=validate, image=%q", image)
	}

	if status != "Success" {
		allowed = false
	}

	admissionResp := &admissionv1beta1.AdmissionResponse{
		UID:     ar.Request.UID,
		Allowed: allowed,
		Result: &metav1.Status{
			Status:  status,
			Message: "ok",
		},
	}

	return admissionResp
}

func (vh *webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	glog.Info("ServeHTTP started")
	// Get webhook body with the admission review.
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		http.Error(w, "no body found", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		fmt.Println(r.URL.Path)
		if r.URL.Path == "/mutate" {
			admissionResponse = vh.mutate(&ar)
		} else if r.URL.Path == "/validate" {
			admissionResponse = vh.validate(&ar)
		} else {
			admissionResponse = &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
	}

	if admissionResponse == nil {
		admissionResponse = &admissionv1beta1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: true,
			Result: &metav1.Status{
				Status:  "Success",
				Message: "ok",
			},
		}
	}
	// Forge the review response.
	aResponse := admissionv1beta1.AdmissionReview{
		Response: admissionResponse,
	}
	resp, err := json.Marshal(aResponse)
	if err != nil {
		http.Error(w, "error marshaling to json admission review response", http.StatusInternalServerError)
		return
	}
	// Forge the HTTP response.
	// If the received admission review has failed mark the response as failed.
	if admissionResponse.Result != nil && admissionResponse.Result.Status == metav1.StatusFailure {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write(resp); err != nil {
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
