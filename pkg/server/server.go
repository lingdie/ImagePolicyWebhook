package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	imagepolicy "k8s.io/api/imagepolicy/v1alpha1"
	"k8s.io/klog/v2"
)

type WebhookServer struct {
	TargetRegistry string
	Server         *http.Server
}

func (whsvr *WebhookServer) Validate(ir *imagepolicy.ImageReview) *imagepolicy.ImageReview {
	return &imagepolicy.ImageReview{
		Status: imagepolicy.ImageReviewStatus{
			Allowed: validateImage(ir, whsvr.TargetRegistry),
		},
	}
}

// getImageRepo returns the repository name of the image
func getImageRepo(image string) (string, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		klog.Errorf("could not parse image reference: %v", err)
		return "", err
	}
	repo := ref.Context().RepositoryStr()
	parts := strings.Split(repo, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid repository name: %s", repo)
	}
	return parts[0], nil
}

// getImageRegistry returns the registry name of the image
func getImageRegistry(image string) (string, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		klog.Errorf("could not parse image reference: %v", err)
		return "", err
	}
	return ref.Context().RegistryStr(), nil
}

func validateImage(ir *imagepolicy.ImageReview, targetRegistry string) bool {
	// check image repository name with image review spec.namespace
	for _, container := range ir.Spec.Containers {
		image := container.Image
		imageRegistry, err := getImageRegistry(image)
		if err != nil {
			klog.Errorf("could not get image registry: %v", err)
			return false
		}
		// if image registry is not target registry, allow it
		if imageRegistry != targetRegistry {
			return true
		}
		imageRepo, err := getImageRepo(image)
		if err != nil {
			klog.Errorf("could not get image repo: %v", err)
			return false
		}
		if imageRepo != ir.Spec.Namespace {
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
