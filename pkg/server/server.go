package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	imagepolicy "k8s.io/api/imagepolicy/v1alpha1"
	"k8s.io/klog/v2"
)

type WebhookServer struct {
	Server *http.Server
}

func (whsvr *WebhookServer) Validate(ir *imagepolicy.ImageReview) *imagepolicy.ImageReview {
	return &imagepolicy.ImageReview{
		Status: imagepolicy.ImageReviewStatus{
			Allowed: validateImage(ir),
		},
	}
}

func validateImage(ir *imagepolicy.ImageReview) bool {
	// validate image, only allow image without "latest" tag
	// TODO: edit this function for your own policy logic
	for _, container := range ir.Spec.Containers {
		// check if image has "latest" tag or without tag, if so, reject
		_, tag, found := strings.Cut(container.Image, ":")
		if !found || tag == "latest" {
			return false
		}
	}
	return true
}

func (whsvr *WebhookServer) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if r.Header.Get("Content-Type") != "application/json" {
		klog.Errorf("invalid Content-Type, expect `application/json`")
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}
	ir := &imagepolicy.ImageReview{}
	if err := json.Unmarshal(body, ir); err != nil {
		klog.Errorf("could not decode body: %v", err)
		http.Error(w, fmt.Sprintf("could not decode body: %v", err), http.StatusBadRequest)
		return
	}
	klog.Infof("validating request: %v", ir.Spec)
	res := whsvr.Validate(ir)
	klog.Infof("validated request %v, with response: %v", ir.Spec, res.Status)
	resp, err := json.Marshal(res)
	if err != nil {
		klog.Errorf("could not encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(resp); err != nil {
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
