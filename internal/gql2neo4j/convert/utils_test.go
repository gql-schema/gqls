package convert

import (
	"testing"
)

func Test_isVersionSupported(t *testing.T) {
	type args struct {
		current string
		minimum string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "empty current version",
			args: args{"", "5.9"},
			want: true,
		},
		{
			name: "current version less than minimum",
			args: args{"5.8", "5.9"},
			want: false,
		},
		{
			name: "current version equal to minimum",
			args: args{"5.9", "5.9"},
			want: true,
		},
		{
			name: "current version greater than minimum",
			args: args{"6.0", "5.9"},
			want: true,
		},
		{
			name: "short new-style version greater than minimum",
			args: args{"2025.01", "5.9"},
			want: true,
		},
		{
			name: "long new-style version greater than minimum",
			args: args{"2025.01.1", "5.9"},
			want: true,
		},
		{
			name:    "invalid current version",
			args:    args{"invalid", "5.9"},
			wantErr: true,
		},
		{
			name:    "invalid minimum version",
			args:    args{"5.9", "invalid"},
			wantErr: true,
		},
		{
			name:    "invalid current and minimum version",
			args:    args{"invalid", "invalid"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isVersionSupported(tt.args.current, tt.args.minimum)
			if (err != nil) != tt.wantErr {
				t.Errorf("isVersionSupported() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isVersionSupported() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mustFormatName(t *testing.T) {
	type args struct {
		prefix     string
		components []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"one component", args{"prefix", []string{"component1"}}, "prefix_component1"},
		{"two components", args{"prefix", []string{"component1", "component2"}}, "prefix_component1_component2"},
		{"uppercase prefix", args{"PREFIX", []string{"component1"}}, "prefix_component1"},
		{"uppercase component", args{"prefix", []string{"COMPONENT1"}}, "prefix_component1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustFormatName(tt.args.prefix, tt.args.components); got != tt.want {
				t.Errorf("mustFormatName() = %v, want %v", got, tt.want)
			}
		})
	}
}
