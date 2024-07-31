package labels

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWasPatchedByKrateo(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected bool
	}{
		{
			name:     "No labels",
			labels:   nil,
			expected: false,
		},
		{
			name: "Label present",
			labels: map[string]string{
				keyPatchedBy: "krateo",
			},
			expected: true,
		},
		{
			name: "Other labels present",
			labels: map[string]string{
				"some-other-label": "some-value",
			},
			expected: false,
		},
		{
			name: "Multiple labels including patched-by",
			labels: map[string]string{
				"some-other-label": "some-value",
				keyPatchedBy:       "krateo",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &corev1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Labels: tt.labels,
				},
			}
			if got := WasPatchedByKrateo(event); got != tt.expected {
				t.Errorf("WasPatchedByKrateo() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCompositionID(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name:     "No labels",
			labels:   nil,
			expected: "",
		},
		{
			name: "Label present",
			labels: map[string]string{
				keyCompositionID: "12345",
			},
			expected: "12345",
		},
		{
			name: "Other labels present",
			labels: map[string]string{
				"some-other-label": "some-value",
			},
			expected: "",
		},
		{
			name: "Multiple labels including composition-id",
			labels: map[string]string{
				"some-other-label": "some-value",
				keyCompositionID:   "67890",
			},
			expected: "67890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &corev1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Labels: tt.labels,
				},
			}
			if got := CompositionID(event); got != tt.expected {
				t.Errorf("CompositionID() = %v, want %v", got, tt.expected)
			}
		})
	}
}
