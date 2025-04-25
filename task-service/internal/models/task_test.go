package models

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestCreateTaskRequest_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		req     CreateTaskRequest
		isValid bool
	}{
		{
			name: "valid input",
			req: CreateTaskRequest{
				Type:   "generate",
				Amount: 10,
				Format: "json",
			},
			isValid: true,
		},
		{
			name: "missing type",
			req: CreateTaskRequest{
				Amount: 5,
				Format: "csv",
			},
			isValid: false,
		},
		{
			name: "amount less than 1",
			req: CreateTaskRequest{
				Type:   "generate",
				Amount: 0,
				Format: "csv",
			},
			isValid: false,
		},
		{
			name: "invalid format",
			req: CreateTaskRequest{
				Type:   "generate",
				Amount: 5,
				Format: "xml",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.req)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
