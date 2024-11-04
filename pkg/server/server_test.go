package server_test

import (
	"testing"

	"github.com/lingdie/image-policy-webhook/pkg/server"
	imagepolicy "k8s.io/api/imagepolicy/v1alpha1"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		images  []string
		allowed bool
	}{
		{
			name:    "deny nginx image",
			images:  []string{"nginx"},
			allowed: false,
		},
		{
			name:    "deny multiple containers",
			images:  []string{"nginx", "redis", "postgres"},
			allowed: false,
		},
		{
			name:    "deny empty image name",
			images:  []string{""},
			allowed: false,
		},
		{
			name:    "allow nginx image",
			images:  []string{"nginx:1.23.4"},
			allowed: true,
		},
	}

	whsvr := &server.WebhookServer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containers := make([]imagepolicy.ImageReviewContainerSpec, len(tt.images))
			for i, img := range tt.images {
				containers[i] = imagepolicy.ImageReviewContainerSpec{Image: img}
			}

			ir := &imagepolicy.ImageReview{
				Spec: imagepolicy.ImageReviewSpec{
					Containers: containers,
				},
			}

			res := whsvr.Validate(ir)
			if res.Status.Allowed != tt.allowed {
				t.Errorf("Validate() got = %v, want %v", res.Status.Allowed, tt.allowed)
			}
		})
	}
}
