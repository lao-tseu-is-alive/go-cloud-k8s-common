package tools

import (
	"testing"
)

func TestToKebabCase(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "it should return a correct kebab case of string",
			args: args{str: "goCloudK8sUserGroup"},
			want: "go-cloud-k8s-user-group",
		},
		{
			name: "it should return a correct kebab case of string with words in Upper",
			args: args{str: "goCloudK8sUserGROUP"},
			want: "go-cloud-k8s-user-group",
		},
		{
			name: "it should return an empty string for a string with space",
			args: args{str: "  "},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToKebabCase(tt.args.str); got != tt.want {
				t.Errorf("ToKebabCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "it should return a correct snake case of string",
			args: args{str: "goCloudK8sUserGroup"},
			want: "go_cloud_k8s_user_group",
		},
		{
			name: "it should return a correct snake case of string with UPPER words ",
			args: args{str: "goCloudK8sUserGROUP"},
			want: "go_cloud_k8s_user_group",
		},
		{
			name: "it should return an empty string for a string with space",
			args: args{str: "  "},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSnakeCase(tt.args.str); got != tt.want {
				t.Errorf("ToSnakeCase() = [%v], want [%v]", got, tt.want)
			}
		})
	}
}
