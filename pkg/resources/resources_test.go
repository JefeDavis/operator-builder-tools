/*
	SPDX-License-Identifier: MIT
*/

package resources

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestEqualGVK(t *testing.T) {
	type args struct {
		left  runtime.Object
		right runtime.Object
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal gvk",
			args: args{
				left: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
				},
				right: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
				},
			},
			want: true,
		},
		{
			name: "inequal gvk (api version)",
			args: args{
				left: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v74",
						Kind:       "Deployment",
					},
				},
				right: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
				},
			},
			want: false,
		},
		{
			name: "inequal gvk (kind)",
			args: args{
				left: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "DaemonSet",
					},
				},
				right: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
				},
			},
			want: false,
		},
		{
			name: "inequal gvk (type)",
			args: args{
				left: &appsv1.DaemonSet{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
				},
				right: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
				},
			},
			want: false,
		},
		{
			name: "inequal gvk (nil)",
			args: args{
				left: &appsv1.DaemonSet{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
				},
				right: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EqualGVK(tt.args.left, tt.args.right); got != tt.want {
				t.Errorf("EqualGVK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEqualNamespaceName(t *testing.T) {
	type args struct {
		left  metav1.Object
		right metav1.Object
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal namespace name",
			args: args{
				left: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-equal",
						Namespace: "test-equal",
					},
				},
				right: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-equal",
						Namespace: "test-equal",
					},
				},
			},
			want: true,
		},
		{
			name: "inequal namespace name (name)",
			args: args{
				left: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-notequal",
						Namespace: "test-equal",
					},
				},
				right: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-equal",
						Namespace: "test-equal",
					},
				},
			},
			want: false,
		},
		{
			name: "inequal namespace name (namespace)",
			args: args{
				left: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-equal",
						Namespace: "test-notequal",
					},
				},
				right: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-equal",
						Namespace: "test-equal",
					},
				},
			},
			want: false,
		},
		{
			name: "inequal namespace name (nil)",
			args: args{
				left: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-notequal",
						Namespace: "test-equal",
					},
				},
				right: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EqualNamespaceName(tt.args.left, tt.args.right); got != tt.want {
				t.Errorf("EqualNamespaceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
