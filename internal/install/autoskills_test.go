package install

import (
	"context"
	"errors"
	"testing"
)

func TestAutoskillsPreflight(t *testing.T) {
	tests := []struct {
		name        string
		nodeVersion string
		nodeErr     error
		npxErr      error
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Success with Node 22.6.0",
			nodeVersion: "v22.6.0\n",
			wantErr:     false,
		},
		{
			name:        "Success with Node 23.0.0",
			nodeVersion: "v23.0.0\n",
			wantErr:     false,
		},
		{
			name:        "Failure with Node 20.0.0",
			nodeVersion: "v20.0.0\n",
			wantErr:     true,
			errMsg:      "Node.js >= 22.6.0 required",
		},
		{
			name:    "Failure when Node is missing",
			nodeErr: errors.New("not found"),
			wantErr: true,
			errMsg:  "Node.js not found",
		},
		{
			name:        "Failure when npx is missing",
			nodeVersion: "v22.6.0\n",
			npxErr:      errors.New("not found"),
			wantErr:     true,
			errMsg:      "npx not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNodeOutput := func() (string, error) {
				return tt.nodeVersion, tt.nodeErr
			}
			mockNpxLookup := func() error {
				return tt.npxErr
			}

			err := autoskillsPreflightWithMocks(mockNodeOutput, mockNpxLookup)
			if (err != nil) != tt.wantErr {
				t.Errorf("AutoskillsPreflight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("AutoskillsPreflight() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestAutoskillsInstall(t *testing.T) {
	tests := []struct {
		name        string
		mockOutput  string
		mockErr     error
		wantSuccess bool
		wantErr     bool
	}{
		{
			name:        "Success",
			mockOutput:  "Detected: React\nInstalled 5 skills\n",
			mockErr:     nil,
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name:        "Failure",
			mockOutput:  "Error: npm registry unreachable\n",
			mockErr:     errors.New("exit status 1"),
			wantSuccess: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := func(ctx context.Context, name string, arg ...string) (string, error) {
				// Verify arguments
				if name != "npx" {
					t.Errorf("expected npx, got %s", name)
				}
				expectedArgs := []string{"-y", "autoskills"}
				for i, argVal := range arg {
					if argVal != expectedArgs[i] {
						t.Errorf("expected arg %s, got %s", expectedArgs[i], argVal)
					}
				}

				return tt.mockOutput, tt.mockErr
			}

			result := autoskillsInstallWithMocks(context.Background(), mockRunner)
			if result.Success != tt.wantSuccess {
				t.Errorf("AutoskillsInstall() success = %v, want %v", result.Success, tt.wantSuccess)
			}
			if (result.Error != nil) != tt.wantErr {
				t.Errorf("AutoskillsInstall() error = %v, wantErr %v", result.Error, tt.wantErr)
			}
			if result.Output != tt.mockOutput {
				t.Errorf("AutoskillsInstall() output = %v, want %v", result.Output, tt.mockOutput)
			}
		})
	}
}
