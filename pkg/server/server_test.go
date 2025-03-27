package server_test

import (
	"testing"

	"github.com/lingdie/image-policy-webhook/pkg/server"
	imagepolicy "k8s.io/api/imagepolicy/v1alpha1"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		images    []string
		allowed   bool
	}{
		{
			name:      "allow nginx image with namespace",
			namespace: "ns-admin",
			images:    []string{"hub.hzh.sealos.run/ns-admin/nginx:1.23.4"},
			allowed:   true,
		},
		{
			name:      "deny nginx image with wrong namespace",
			namespace: "ns-admin",
			images:    []string{"hub.hzh.sealos.run/ns-notadmin/nginx:1.23.4"},
			allowed:   false,
		},
		{
			name:      "allow docker.io nginx image",
			namespace: "ns-admin",
			images:    []string{"nginx:1.23.4"},
			allowed:   true,
		},
		{
			name:      "deny empty image name",
			namespace: "ns-admin",
			images:    []string{""},
			allowed:   false,
		},
	}

	whsvr := &server.WebhookServer{
		TargetRegistry: "hub.hzh.sealos.run",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containers := make([]imagepolicy.ImageReviewContainerSpec, len(tt.images))
			for i, img := range tt.images {
				containers[i] = imagepolicy.ImageReviewContainerSpec{Image: img}
			}

			ir := &imagepolicy.ImageReview{
				Spec: imagepolicy.ImageReviewSpec{
					Containers: containers,
					Namespace:  tt.namespace,
				},
			}

			res := whsvr.Validate(ir)
			if res.Status.Allowed != tt.allowed {
				t.Errorf("Validate() got = %v, want %v", res.Status.Allowed, tt.allowed)
			}
		})
	}
}
