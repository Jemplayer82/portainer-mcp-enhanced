package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jmrplens/portainer-mcp-enhanced/pkg/portainer/models"
)

// TestHandleGetStacks verifies the HandleGetStacks MCP tool handler.
func TestHandleGetStacks(t *testing.T) {
	tests := []struct {
		name        string
		mockStacks  []models.Stack
		mockError   error
		expectError bool
	}{
		{
			name: "successful stacks retrieval",
			mockStacks: []models.Stack{
				{ID: 1, Name: "stack1"},
				{ID: 2, Name: "stack2"},
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "api error",
			mockStacks:  nil,
			mockError:   fmt.Errorf("api error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			mockClient.On("GetStacks").Return(tt.mockStacks, tt.mockError)

			server := &PortainerMCPServer{
				cli: mockClient,
			}

			handler := server.HandleGetStacks()
			result, err := handler(context.Background(), mcp.CallToolRequest{})

			if tt.expectError {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError, "result.IsError should be true for expected errors")
				assert.Len(t, result.Content, 1)
				textContent, ok := result.Content[0].(mcp.TextContent)
				assert.True(t, ok, "Result content should be mcp.TextContent for errors")
				if tt.mockError != nil {
					assert.Contains(t, textContent.Text, tt.mockError.Error())
				} else {
					assert.NotEmpty(t, textContent.Text, "Error message should not be empty for parameter errors")
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, result.Content, 1)
				textContent, ok := result.Content[0].(mcp.TextContent)
				assert.True(t, ok)

				var stacks []models.Stack
				err = json.Unmarshal([]byte(textContent.Text), &stacks)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockStacks, stacks)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleGetStackFile verifies the HandleGetStackFile MCP tool handler.
func TestHandleGetStackFile(t *testing.T) {
	tests := []struct {
		name        string
		inputID     int
		mockContent string
		mockError   error
		expectError bool
		setupParams func(request *mcp.CallToolRequest)
	}{
		{
			name:        "successful file retrieval",
			inputID:     1,
			mockContent: "version: '3'\nservices:\n  web:\n    image: nginx",
			mockError:   nil,
			expectError: false,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"id": float64(1),
				}
			},
		},
		{
			name:        "api error",
			inputID:     1,
			mockContent: "",
			mockError:   fmt.Errorf("api error"),
			expectError: true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"id": float64(1),
				}
			},
		},
		{
			name:        "missing id parameter",
			inputID:     0,
			mockContent: "",
			mockError:   nil,
			expectError: true,
			setupParams: func(request *mcp.CallToolRequest) {
				// No need to set any parameters as the request will be invalid
			},
		},
		{
			name:        "invalid id zero",
			inputID:     0,
			mockContent: "",
			mockError:   nil,
			expectError: true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"id": float64(0),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			if !tt.expectError || tt.mockError != nil {
				mockClient.On("GetStackFile", tt.inputID).Return(tt.mockContent, tt.mockError)
			}

			server := &PortainerMCPServer{
				cli: mockClient,
			}

			request := CreateMCPRequest(map[string]any{})
			tt.setupParams(&request)

			handler := server.HandleGetStackFile()
			result, err := handler(context.Background(), request)

			if tt.expectError {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError, "result.IsError should be true for expected errors")
				assert.Len(t, result.Content, 1)
				textContent, ok := result.Content[0].(mcp.TextContent)
				assert.True(t, ok, "Result content should be mcp.TextContent for errors")
				if tt.mockError != nil {
					assert.Contains(t, textContent.Text, tt.mockError.Error())
				} else {
					assert.NotEmpty(t, textContent.Text, "Error message should not be empty for parameter errors")
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, result.Content, 1)
				textContent, ok := result.Content[0].(mcp.TextContent)
				assert.True(t, ok)
				assert.Equal(t, tt.mockContent, textContent.Text)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleCreateStack verifies the HandleCreateStack MCP tool handler.
func TestHandleCreateStack(t *testing.T) {
	tests := []struct {
		name             string
		inputName        string
		inputFile        string
		inputEnvGroupIDs []int
		mockID           int
		mockError        error
		expectError      bool
		setupParams      func(request *mcp.CallToolRequest)
	}{
		{
			name:             "successful stack creation",
			inputName:        "test-stack",
			inputFile:        "version: '3'\nservices:\n  web:\n    image: nginx",
			inputEnvGroupIDs: []int{1, 2},
			mockID:           1,
			mockError:        nil,
			expectError:      false,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"name":                "test-stack",
					"file":                "version: '3'\nservices:\n  web:\n    image: nginx",
					"environmentGroupIds": []any{float64(1), float64(2)},
				}
			},
		},
		{
			name:             "api error",
			inputName:        "test-stack",
			inputFile:        "version: '3'\nservices:\n  web:\n    image: nginx",
			inputEnvGroupIDs: []int{1, 2},
			mockID:           0,
			mockError:        fmt.Errorf("api error"),
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"name":                "test-stack",
					"file":                "version: '3'\nservices:\n  web:\n    image: nginx",
					"environmentGroupIds": []any{float64(1), float64(2)},
				}
			},
		},
		{
			name:             "missing name parameter",
			inputName:        "",
			inputFile:        "version: '3'\nservices:\n  web:\n    image: nginx",
			inputEnvGroupIDs: []int{1, 2},
			mockID:           0,
			mockError:        nil,
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"file":                "version: '3'\nservices:\n  web:\n    image: nginx",
					"environmentGroupIds": []any{float64(1), float64(2)},
				}
			},
		},
		{
			name:             "missing file parameter",
			inputName:        "test-stack",
			inputFile:        "",
			inputEnvGroupIDs: []int{1, 2},
			mockID:           0,
			mockError:        nil,
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"name":                "test-stack",
					"environmentGroupIds": []any{float64(1), float64(2)},
				}
			},
		},
		{
			name:             "missing environmentGroupIds parameter",
			inputName:        "test-stack",
			inputFile:        "version: '3'\nservices:\n  web:\n    image: nginx",
			inputEnvGroupIDs: nil,
			mockID:           0,
			mockError:        nil,
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"name": "test-stack",
					"file": "version: '3'\nservices:\n  web:\n    image: nginx",
				}
			},
		},
		{
			name:             "empty name triggers validateName error",
			inputName:        "",
			inputFile:        "version: '3'\nservices:\n  web:\n    image: nginx",
			inputEnvGroupIDs: []int{1},
			mockID:           0,
			mockError:        nil,
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"name":                "   ",
					"file":                "version: '3'\nservices:\n  web:\n    image: nginx",
					"environmentGroupIds": []any{float64(1)},
				}
			},
		},
		{
			name:             "invalid YAML file triggers validateComposeYAML error",
			inputName:        "test-stack",
			inputFile:        "",
			inputEnvGroupIDs: []int{1},
			mockID:           0,
			mockError:        nil,
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"name":                "test-stack",
					"file":                ":\ninvalid: {{{yaml",
					"environmentGroupIds": []any{float64(1)},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			if !tt.expectError || tt.mockError != nil {
				mockClient.On("CreateStack", tt.inputName, tt.inputFile, tt.inputEnvGroupIDs).Return(tt.mockID, tt.mockError)
			}

			server := &PortainerMCPServer{
				cli: mockClient,
			}

			request := CreateMCPRequest(map[string]any{})
			tt.setupParams(&request)

			handler := server.HandleCreateStack()
			result, err := handler(context.Background(), request)

			if tt.expectError {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError, "result.IsError should be true for expected errors")
				assert.Len(t, result.Content, 1)
				textContent, ok := result.Content[0].(mcp.TextContent)
				assert.True(t, ok, "Result content should be mcp.TextContent for errors")
				if tt.mockError != nil {
					assert.Contains(t, textContent.Text, tt.mockError.Error())
				} else {
					assert.NotEmpty(t, textContent.Text, "Error message should not be empty for parameter errors")
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, result.Content, 1)
				textContent, ok := result.Content[0].(mcp.TextContent)
				assert.True(t, ok)
				assert.Contains(t, textContent.Text, fmt.Sprintf("ID: %d", tt.mockID))
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleUpdateStack verifies the HandleUpdateStack MCP tool handler.
func TestHandleUpdateStack(t *testing.T) {
	tests := []struct {
		name             string
		inputID          int
		inputFile        string
		inputEnvGroupIDs []int
		mockError        error
		expectError      bool
		setupParams      func(request *mcp.CallToolRequest)
	}{
		{
			name:             "successful stack update",
			inputID:          1,
			inputFile:        "version: '3'\nservices:\n  web:\n    image: nginx",
			inputEnvGroupIDs: []int{1, 2},
			mockError:        nil,
			expectError:      false,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"id":                  float64(1),
					"file":                "version: '3'\nservices:\n  web:\n    image: nginx",
					"environmentGroupIds": []any{float64(1), float64(2)},
				}
			},
		},
		{
			name:             "api error",
			inputID:          1,
			inputFile:        "version: '3'\nservices:\n  web:\n    image: nginx",
			inputEnvGroupIDs: []int{1, 2},
			mockError:        fmt.Errorf("api error"),
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"id":                  float64(1),
					"file":                "version: '3'\nservices:\n  web:\n    image: nginx",
					"environmentGroupIds": []any{float64(1), float64(2)},
				}
			},
		},
		{
			name:             "missing id parameter",
			inputID:          0,
			inputFile:        "version: '3'\nservices:\n  web:\n    image: nginx",
			inputEnvGroupIDs: []int{1, 2},
			mockError:        nil,
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"file":                "version: '3'\nservices:\n  web:\n    image: nginx",
					"environmentGroupIds": []any{float64(1), float64(2)},
				}
			},
		},
		{
			name:             "missing file parameter",
			inputID:          1,
			inputFile:        "",
			inputEnvGroupIDs: []int{1, 2},
			mockError:        nil,
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"id":                  float64(1),
					"environmentGroupIds": []any{float64(1), float64(2)},
				}
			},
		},
		{
			name:             "missing environmentGroupIds parameter",
			inputID:          1,
			inputFile:        "version: '3'\nservices:\n  web:\n    image: nginx",
			inputEnvGroupIDs: nil,
			mockError:        nil,
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"id":   float64(1),
					"file": "version: '3'\nservices:\n  web:\n    image: nginx",
				}
			},
		},
		{
			name:             "invalid id zero triggers validatePositiveID error",
			inputID:          0,
			inputFile:        "version: '3'\nservices:\n  web:\n    image: nginx",
			inputEnvGroupIDs: []int{1},
			mockError:        nil,
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"id":                  float64(0),
					"file":                "version: '3'\nservices:\n  web:\n    image: nginx",
					"environmentGroupIds": []any{float64(1)},
				}
			},
		},
		{
			name:             "invalid YAML file triggers validateComposeYAML error",
			inputID:          1,
			inputFile:        "",
			inputEnvGroupIDs: []int{1},
			mockError:        nil,
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"id":                  float64(1),
					"file":                ":\ninvalid: {{{yaml",
					"environmentGroupIds": []any{float64(1)},
				}
			},
		},
		{
			name:             "wrong type for environmentGroupIds triggers GetArrayOfIntegers error",
			inputID:          1,
			inputFile:        "version: '3'\nservices:\n  web:\n    image: nginx",
			inputEnvGroupIDs: nil,
			mockError:        nil,
			expectError:      true,
			setupParams: func(request *mcp.CallToolRequest) {
				request.Params.Arguments = map[string]any{
					"id":                  float64(1),
					"file":                "version: '3'\nservices:\n  web:\n    image: nginx",
					"environmentGroupIds": "not-an-array",
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			if !tt.expectError || tt.mockError != nil {
				mockClient.On("UpdateStack", tt.inputID, tt.inputFile, tt.inputEnvGroupIDs).Return(tt.mockError)
			}

			server := &PortainerMCPServer{
				cli: mockClient,
			}

			request := CreateMCPRequest(map[string]any{})
			tt.setupParams(&request)

			handler := server.HandleUpdateStack()
			result, err := handler(context.Background(), request)

			if tt.expectError {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError, "result.IsError should be true for expected errors")
				assert.Len(t, result.Content, 1)
				textContent, ok := result.Content[0].(mcp.TextContent)
				assert.True(t, ok, "Result content should be mcp.TextContent for errors")
				if tt.mockError != nil {
					assert.Contains(t, textContent.Text, tt.mockError.Error())
				} else {
					assert.NotEmpty(t, textContent.Text, "Error message should not be empty for parameter errors")
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, result.Content, 1)
				textContent, ok := result.Content[0].(mcp.TextContent)
				assert.True(t, ok)
				assert.Contains(t, textContent.Text, "successfully")
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleListRegularStacks verifies the HandleListRegularStacks MCP tool handler.
func TestHandleListRegularStacks(t *testing.T) {
	tests := []struct {
		name        string
		mockStacks  []models.RegularStack
		mockError   error
		expectError bool
	}{
		{
			name: "successful regular stacks retrieval",
			mockStacks: []models.RegularStack{
				{ID: 1, Name: "web-app", Status: 1, EndpointID: 2},
				{ID: 2, Name: "db-stack", Status: 1, EndpointID: 3},
			},
			expectError: false,
		},
		{
			name:        "empty list",
			mockStacks:  []models.RegularStack{},
			expectError: false,
		},
		{
			name:        "api error",
			mockError:   fmt.Errorf("connection refused"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			mockClient.On("GetRegularStacks").Return(tt.mockStacks, tt.mockError)

			s := &PortainerMCPServer{cli: mockClient}
			handler := s.HandleListRegularStacks()
			result, err := handler(context.Background(), mcp.CallToolRequest{})

			assert.NoError(t, err)
			if tt.expectError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
				var stacks []models.RegularStack
				textContent := result.Content[0].(mcp.TextContent)
				unmarshalErr := json.Unmarshal([]byte(textContent.Text), &stacks)
				assert.NoError(t, unmarshalErr)
				assert.Equal(t, len(tt.mockStacks), len(stacks))
			}
			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleInspectStack verifies the HandleInspectStack MCP tool handler.
func TestHandleInspectStack(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		mockStack   models.RegularStack
		mockError   error
		expectError bool
	}{
		{
			name:      "successful inspect",
			params:    map[string]any{"id": float64(1)},
			mockStack: models.RegularStack{ID: 1, Name: "my-stack", Status: 1},
		},
		{
			name:        "missing id",
			params:      map[string]any{},
			expectError: true,
		},
		{
			name:        "invalid id zero",
			params:      map[string]any{"id": float64(0)},
			expectError: true,
		},
		{
			name:        "negative id",
			params:      map[string]any{"id": float64(-1)},
			expectError: true,
		},
		{
			name:        "api error",
			params:      map[string]any{"id": float64(1)},
			mockError:   fmt.Errorf("not found"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			if idVal, ok := tt.params["id"]; ok && idVal.(float64) > 0 {
				mockClient.On("InspectStack", int(idVal.(float64))).Return(tt.mockStack, tt.mockError)
			}

			s := &PortainerMCPServer{cli: mockClient}
			handler := s.HandleInspectStack()
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params
			result, err := handler(context.Background(), req)

			assert.NoError(t, err)
			if tt.expectError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
				var stack models.RegularStack
				textContent := result.Content[0].(mcp.TextContent)
				unmarshalErr := json.Unmarshal([]byte(textContent.Text), &stack)
				assert.NoError(t, unmarshalErr)
				assert.Equal(t, tt.mockStack.ID, stack.ID)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleDeleteStack verifies the HandleDeleteStack MCP tool handler.
func TestHandleDeleteStack(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		mockError   error
		expectError bool
	}{
		{
			name:   "successful delete",
			params: map[string]any{"id": float64(1), "environmentId": float64(2), "removeVolumes": true},
		},
		{
			name:   "successful delete without removeVolumes",
			params: map[string]any{"id": float64(1), "environmentId": float64(2)},
		},
		{
			name:        "missing id",
			params:      map[string]any{"environmentId": float64(2)},
			expectError: true,
		},
		{
			name:        "missing environmentId",
			params:      map[string]any{"id": float64(1)},
			expectError: true,
		},
		{
			name:        "invalid id zero",
			params:      map[string]any{"id": float64(0), "environmentId": float64(2)},
			expectError: true,
		},
		{
			name:        "invalid environmentId zero",
			params:      map[string]any{"id": float64(1), "environmentId": float64(0)},
			expectError: true,
		},
		{
			name:        "invalid removeVolumes type triggers GetBoolean error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2), "removeVolumes": "not-a-bool"},
			expectError: true,
		},
		{
			name:        "api error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2)},
			mockError:   fmt.Errorf("forbidden"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			idVal, hasID := tt.params["id"]
			envVal, hasEnv := tt.params["environmentId"]
			_, removeVolumesIsInvalid := tt.params["removeVolumes"].(string)
			if hasID && hasEnv && idVal.(float64) > 0 && envVal.(float64) > 0 && !removeVolumesIsInvalid {
				removeVolumes, _ := tt.params["removeVolumes"].(bool)
				mockClient.On("DeleteStack", int(idVal.(float64)), models.DeleteStackOptions{
					EndpointID:    int(envVal.(float64)),
					RemoveVolumes: removeVolumes,
				}).Return(tt.mockError)
			}

			s := &PortainerMCPServer{cli: mockClient}
			handler := s.HandleDeleteStack()
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params
			result, err := handler(context.Background(), req)

			assert.NoError(t, err)
			if tt.expectError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
				textContent := result.Content[0].(mcp.TextContent)
				assert.Contains(t, textContent.Text, "successfully")
			}
			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleInspectStackFile verifies the HandleInspectStackFile MCP tool handler.
func TestHandleInspectStackFile(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		mockContent string
		mockError   error
		expectError bool
	}{
		{
			name:        "successful file retrieval",
			params:      map[string]any{"id": float64(1)},
			mockContent: "version: '3'\nservices:\n  web:\n    image: nginx",
		},
		{
			name:        "missing id",
			params:      map[string]any{},
			expectError: true,
		},
		{
			name:        "invalid id",
			params:      map[string]any{"id": float64(0)},
			expectError: true,
		},
		{
			name:        "api error",
			params:      map[string]any{"id": float64(1)},
			mockError:   fmt.Errorf("not found"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			if idVal, ok := tt.params["id"]; ok && idVal.(float64) > 0 {
				mockClient.On("InspectStackFile", int(idVal.(float64))).Return(tt.mockContent, tt.mockError)
			}

			s := &PortainerMCPServer{cli: mockClient}
			handler := s.HandleInspectStackFile()
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params
			result, err := handler(context.Background(), req)

			assert.NoError(t, err)
			if tt.expectError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
				textContent := result.Content[0].(mcp.TextContent)
				assert.Equal(t, tt.mockContent, textContent.Text)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleUpdateStackGit verifies the HandleUpdateStackGit MCP tool handler.
func TestHandleUpdateStackGit(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		mockStack   models.RegularStack
		mockError   error
		expectError bool
	}{
		{
			name:      "successful update with all params",
			params:    map[string]any{"id": float64(1), "environmentId": float64(2), "referenceName": "main", "prune": true},
			mockStack: models.RegularStack{ID: 1, Name: "my-stack"},
		},
		{
			name:      "successful update with minimal params",
			params:    map[string]any{"id": float64(1), "environmentId": float64(2)},
			mockStack: models.RegularStack{ID: 1, Name: "my-stack"},
		},
		{
			name:        "missing id",
			params:      map[string]any{"environmentId": float64(2)},
			expectError: true,
		},
		{
			name:        "missing environmentId",
			params:      map[string]any{"id": float64(1)},
			expectError: true,
		},
		{
			name:        "invalid id",
			params:      map[string]any{"id": float64(0), "environmentId": float64(2)},
			expectError: true,
		},
		{
			name:        "invalid environmentId",
			params:      map[string]any{"id": float64(1), "environmentId": float64(-1)},
			expectError: true,
		},
		{
			name:        "invalid referenceName type triggers GetString error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2), "referenceName": float64(42)},
			expectError: true,
		},
		{
			name:        "invalid prune type triggers GetBoolean error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2), "prune": "not-a-bool"},
			expectError: true,
		},
		{
			name:        "api error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2)},
			mockError:   fmt.Errorf("conflict"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			idVal, hasID := tt.params["id"]
			envVal, hasEnv := tt.params["environmentId"]
			_, refNameInvalid := tt.params["referenceName"].(float64)
			_, pruneInvalid := tt.params["prune"].(string)
			if hasID && hasEnv && idVal.(float64) > 0 && envVal.(float64) > 0 && !refNameInvalid && !pruneInvalid {
				refName, _ := tt.params["referenceName"].(string)
				prune, _ := tt.params["prune"].(bool)
				mockClient.On("UpdateStackGit", int(idVal.(float64)), models.UpdateStackGitOptions{
					EndpointID:    int(envVal.(float64)),
					ReferenceName: refName,
					Prune:         prune,
				}).Return(tt.mockStack, tt.mockError)
			}

			s := &PortainerMCPServer{cli: mockClient}
			handler := s.HandleUpdateStackGit()
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params
			result, err := handler(context.Background(), req)

			assert.NoError(t, err)
			if tt.expectError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleRedeployStackGit verifies the HandleRedeployStackGit MCP tool handler.
func TestHandleRedeployStackGit(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		mockStack   models.RegularStack
		mockError   error
		expectError bool
	}{
		{
			// pullImage:true is covered separately by TestHandleRedeployStackGit_ForcePull,
			// which also mocks the force-pull calls this now triggers.
			name:      "successful redeploy with prune",
			params:    map[string]any{"id": float64(1), "environmentId": float64(2), "pullImage": false, "prune": true},
			mockStack: models.RegularStack{ID: 1, Name: "redeployed"},
		},
		{
			name:      "successful redeploy minimal",
			params:    map[string]any{"id": float64(1), "environmentId": float64(2)},
			mockStack: models.RegularStack{ID: 1, Name: "redeployed"},
		},
		{
			name:        "missing id",
			params:      map[string]any{"environmentId": float64(2)},
			expectError: true,
		},
		{
			name:        "missing environmentId",
			params:      map[string]any{"id": float64(1)},
			expectError: true,
		},
		{
			name:        "invalid id",
			params:      map[string]any{"id": float64(0), "environmentId": float64(2)},
			expectError: true,
		},
		{
			name:        "invalid environmentId zero triggers validatePositiveID error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(0)},
			expectError: true,
		},
		{
			name:        "invalid pullImage type triggers GetBoolean error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2), "pullImage": "not-a-bool"},
			expectError: true,
		},
		{
			name:        "invalid prune type triggers GetBoolean error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2), "prune": "not-a-bool"},
			expectError: true,
		},
		{
			name:        "api error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2)},
			mockError:   fmt.Errorf("deploy error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			idVal, hasID := tt.params["id"]
			envVal, hasEnv := tt.params["environmentId"]
			_, pullImageInvalid := tt.params["pullImage"].(string)
			_, pruneInvalid := tt.params["prune"].(string)
			if hasID && hasEnv && idVal.(float64) > 0 && envVal.(float64) > 0 && !pullImageInvalid && !pruneInvalid {
				pullImage, _ := tt.params["pullImage"].(bool)
				prune, _ := tt.params["prune"].(bool)
				mockClient.On("RedeployStackGit", int(idVal.(float64)), models.RedeployStackGitOptions{
					EndpointID: int(envVal.(float64)),
					PullImage:  pullImage,
					Prune:      prune,
				}).Return(tt.mockStack, tt.mockError)
			}

			s := &PortainerMCPServer{cli: mockClient}
			handler := s.HandleRedeployStackGit()
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params
			result, err := handler(context.Background(), req)

			assert.NoError(t, err)
			if tt.expectError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

// fakeImageCreateResponse builds an *http.Response resembling a Docker Engine
// POST /images/create streamed response, for mocking ProxyDockerRequest.
func fakeImageCreateResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// TestHandleRedeployStackGit_ForcePull verifies the force-pull behavior that
// HandleRedeployStackGit performs ahead of the actual redeploy when
// pullImage=true: reading the stack file, pulling each mutable-tag image via
// the Docker proxy, and surfacing the outcome in the response without ever
// blocking the redeploy itself.
func TestHandleRedeployStackGit_ForcePull(t *testing.T) {
	const composeOneImage = "services:\n  web:\n    image: ghcr.io/example/app:latest\n"
	const composeTwoImages = "services:\n  web:\n    image: ghcr.io/example/app:latest\n  db:\n    image: postgres:16\n  worker:\n    build: .\n  pinned:\n    image: redis@sha256:deadbeef\n"

	params := map[string]any{"id": float64(1), "environmentId": float64(2), "pullImage": true}
	redeployOpts := models.RedeployStackGitOptions{EndpointID: 2, PullImage: true, Prune: false}
	mockStack := models.RegularStack{ID: 1, Name: "redeployed"}

	t.Run("successful pull is reported and redeploy proceeds", func(t *testing.T) {
		mockClient := &MockPortainerClient{}
		mockClient.On("InspectStackFile", 1).Return(composeOneImage, nil)
		mockClient.On("ProxyDockerRequest", mock.MatchedBy(func(opts models.DockerProxyRequestOptions) bool {
			return opts.EnvironmentID == 2 && opts.Method == http.MethodPost && opts.Path == "/images/create" &&
				opts.QueryParams["fromImage"] == "ghcr.io/example/app" && opts.QueryParams["tag"] == "latest"
		})).Return(fakeImageCreateResponse(http.StatusOK, `{"status":"Pulling from example/app"}`+"\n"+`{"status":"Status: Downloaded newer image for ghcr.io/example/app:latest"}`), nil)
		mockClient.On("RedeployStackGit", 1, redeployOpts).Return(mockStack, nil)

		s := &PortainerMCPServer{cli: mockClient}
		req := mcp.CallToolRequest{}
		req.Params.Arguments = params
		result, err := s.HandleRedeployStackGit()(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, result.IsError)

		var resp redeployStackGitResponse
		assert.NoError(t, json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &resp))
		assert.Equal(t, mockStack.ID, resp.ID)
		if assert.Len(t, resp.ImagePulls, 1) {
			assert.Equal(t, "ghcr.io/example/app:latest", resp.ImagePulls[0].Image)
			assert.True(t, resp.ImagePulls[0].Pulled)
			assert.Equal(t, "Status: Downloaded newer image for ghcr.io/example/app:latest", resp.ImagePulls[0].Message)
		}
		mockClient.AssertExpectations(t)
	})

	t.Run("pull failure is reported but does not block redeploy", func(t *testing.T) {
		mockClient := &MockPortainerClient{}
		mockClient.On("InspectStackFile", 1).Return(composeOneImage, nil)
		mockClient.On("ProxyDockerRequest", mock.Anything).
			Return(fakeImageCreateResponse(http.StatusNotFound, `{"error":"pull access denied for ghcr.io/example/app"}`), nil)
		mockClient.On("RedeployStackGit", 1, redeployOpts).Return(mockStack, nil)

		s := &PortainerMCPServer{cli: mockClient}
		req := mcp.CallToolRequest{}
		req.Params.Arguments = params
		result, err := s.HandleRedeployStackGit()(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, result.IsError, "a failed force-pull must not fail the whole redeploy")

		var resp redeployStackGitResponse
		assert.NoError(t, json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &resp))
		if assert.Len(t, resp.ImagePulls, 1) {
			assert.False(t, resp.ImagePulls[0].Pulled)
			assert.Contains(t, resp.ImagePulls[0].Message, "pull access denied")
		}
		mockClient.AssertExpectations(t)
	})

	t.Run("stack file read failure falls back to old behavior", func(t *testing.T) {
		mockClient := &MockPortainerClient{}
		mockClient.On("InspectStackFile", 1).Return("", fmt.Errorf("not found"))
		mockClient.On("RedeployStackGit", 1, redeployOpts).Return(mockStack, nil)

		s := &PortainerMCPServer{cli: mockClient}
		req := mcp.CallToolRequest{}
		req.Params.Arguments = params
		result, err := s.HandleRedeployStackGit()(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, result.IsError)

		// No imagePulls key at all when the stack file couldn't be read --
		// ProxyDockerRequest must never be called in this case.
		assert.NotContains(t, result.Content[0].(mcp.TextContent).Text, "imagePulls")
		mockClient.AssertExpectations(t)
	})

	t.Run("multiple services: build-only and digest-pinned images are skipped", func(t *testing.T) {
		mockClient := &MockPortainerClient{}
		mockClient.On("InspectStackFile", 1).Return(composeTwoImages, nil)
		mockClient.On("ProxyDockerRequest", mock.Anything).
			Return(fakeImageCreateResponse(http.StatusOK, `{"status":"Status: Image is up to date"}`), nil)
		mockClient.On("RedeployStackGit", 1, redeployOpts).Return(mockStack, nil)

		s := &PortainerMCPServer{cli: mockClient}
		req := mcp.CallToolRequest{}
		req.Params.Arguments = params
		result, err := s.HandleRedeployStackGit()(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, result.IsError)

		var resp redeployStackGitResponse
		assert.NoError(t, json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &resp))
		// Only "web" and "db" have a mutable-tag image; "worker" builds locally
		// and "pinned" is already digest-pinned, so neither triggers a pull.
		assert.Len(t, resp.ImagePulls, 2)
		mockClient.AssertExpectations(t)
	})
}

// TestExtractComposeImages verifies compose YAML image-reference extraction.
func TestExtractComposeImages(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "single service with image",
			content: "services:\n  web:\n    image: nginx:latest\n",
			want:    []string{"nginx:latest"},
		},
		{
			name:    "build-only service is skipped",
			content: "services:\n  worker:\n    build: .\n",
			want:    []string{},
		},
		{
			name:    "digest-pinned image is excluded",
			content: "services:\n  web:\n    image: nginx@sha256:abcdef1234567890\n",
			want:    []string{},
		},
		{
			name:    "invalid yaml yields no images",
			content: "not: valid: yaml: [",
			want:    nil,
		},
		{
			name:    "empty content yields no images",
			content: "",
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractComposeImages(tt.content)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

// TestSplitImageRef verifies image reference parsing into repo and tag.
func TestSplitImageRef(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		wantRepo string
		wantTag  string
	}{
		{name: "no tag defaults to latest", image: "redis", wantRepo: "redis", wantTag: "latest"},
		{name: "explicit tag", image: "redis:7", wantRepo: "redis", wantTag: "7"},
		{name: "namespaced repo with tag", image: "ghcr.io/example/app:v1.2.3", wantRepo: "ghcr.io/example/app", wantTag: "v1.2.3"},
		{name: "registry host:port without tag", image: "myregistry:5000/repo", wantRepo: "myregistry:5000/repo", wantTag: "latest"},
		{name: "registry host:port with tag", image: "myregistry:5000/repo:tag", wantRepo: "myregistry:5000/repo", wantTag: "tag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, tag := splitImageRef(tt.image)
			assert.Equal(t, tt.wantRepo, repo)
			assert.Equal(t, tt.wantTag, tag)
		})
	}
}

// TestReadImageCreateStatus verifies parsing of the Docker Engine
// POST /images/create streamed response.
func TestReadImageCreateStatus(t *testing.T) {
	t.Run("success stream returns last status", func(t *testing.T) {
		body := `{"status":"Pulling from library/redis"}` + "\n" + `{"status":"Status: Downloaded newer image for redis:latest"}`
		status, err := readImageCreateStatus(fakeImageCreateResponse(http.StatusOK, body))
		assert.NoError(t, err)
		assert.Equal(t, "Status: Downloaded newer image for redis:latest", status)
	})

	t.Run("error field in stream is surfaced", func(t *testing.T) {
		body := `{"errorDetail":{"message":"pull access denied"},"error":"pull access denied"}`
		status, err := readImageCreateStatus(fakeImageCreateResponse(http.StatusOK, body))
		assert.Error(t, err)
		assert.Contains(t, status, "pull access denied")
	})

	t.Run("non-2xx status without an error field is treated as failure", func(t *testing.T) {
		_, err := readImageCreateStatus(fakeImageCreateResponse(http.StatusUnauthorized, `{"message":"unauthorized"}`))
		assert.Error(t, err)
	})

	t.Run("garbage lines are ignored", func(t *testing.T) {
		body := "not json\n" + `{"status":"Status: Image is up to date"}` + "\nmore garbage"
		status, err := readImageCreateStatus(fakeImageCreateResponse(http.StatusOK, body))
		assert.NoError(t, err)
		assert.Equal(t, "Status: Image is up to date", status)
	})

	t.Run("empty body with 2xx status is an error", func(t *testing.T) {
		_, err := readImageCreateStatus(fakeImageCreateResponse(http.StatusOK, ""))
		assert.Error(t, err)
	})
}

// TestHandleStartStack verifies the HandleStartStack MCP tool handler.
func TestHandleStartStack(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		mockStack   models.RegularStack
		mockError   error
		expectError bool
	}{
		{
			name:      "successful start",
			params:    map[string]any{"id": float64(1), "environmentId": float64(2)},
			mockStack: models.RegularStack{ID: 1, Name: "started-stack", Status: 1},
		},
		{
			name:        "missing id",
			params:      map[string]any{"environmentId": float64(2)},
			expectError: true,
		},
		{
			name:        "missing environmentId",
			params:      map[string]any{"id": float64(1)},
			expectError: true,
		},
		{
			name:        "invalid id",
			params:      map[string]any{"id": float64(-5), "environmentId": float64(2)},
			expectError: true,
		},
		{
			name:        "invalid environmentId",
			params:      map[string]any{"id": float64(1), "environmentId": float64(0)},
			expectError: true,
		},
		{
			name:        "api error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2)},
			mockError:   fmt.Errorf("start failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			idVal, hasID := tt.params["id"]
			envVal, hasEnv := tt.params["environmentId"]
			if hasID && hasEnv && idVal.(float64) > 0 && envVal.(float64) > 0 {
				mockClient.On("StartStack", int(idVal.(float64)), int(envVal.(float64))).Return(tt.mockStack, tt.mockError)
			}

			s := &PortainerMCPServer{cli: mockClient}
			handler := s.HandleStartStack()
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params
			result, err := handler(context.Background(), req)

			assert.NoError(t, err)
			if tt.expectError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleStopStack verifies the HandleStopStack MCP tool handler.
func TestHandleStopStack(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		mockStack   models.RegularStack
		mockError   error
		expectError bool
	}{
		{
			name:      "successful stop",
			params:    map[string]any{"id": float64(1), "environmentId": float64(2)},
			mockStack: models.RegularStack{ID: 1, Name: "stopped-stack", Status: 2},
		},
		{
			name:        "missing id",
			params:      map[string]any{"environmentId": float64(2)},
			expectError: true,
		},
		{
			name:        "missing environmentId",
			params:      map[string]any{"id": float64(1)},
			expectError: true,
		},
		{
			name:        "invalid id",
			params:      map[string]any{"id": float64(0), "environmentId": float64(2)},
			expectError: true,
		},
		{
			name:        "invalid environmentId zero triggers validatePositiveID error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(0)},
			expectError: true,
		},
		{
			name:        "api error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2)},
			mockError:   fmt.Errorf("stop failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			idVal, hasID := tt.params["id"]
			envVal, hasEnv := tt.params["environmentId"]
			if hasID && hasEnv && idVal.(float64) > 0 && envVal.(float64) > 0 {
				mockClient.On("StopStack", int(idVal.(float64)), int(envVal.(float64))).Return(tt.mockStack, tt.mockError)
			}

			s := &PortainerMCPServer{cli: mockClient}
			handler := s.HandleStopStack()
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params
			result, err := handler(context.Background(), req)

			assert.NoError(t, err)
			if tt.expectError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleMigrateStack verifies the HandleMigrateStack MCP tool handler.
func TestHandleMigrateStack(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		mockStack   models.RegularStack
		mockError   error
		expectError bool
	}{
		{
			name:      "successful migrate with name",
			params:    map[string]any{"id": float64(1), "environmentId": float64(2), "targetEnvironmentId": float64(3), "name": "new-name"},
			mockStack: models.RegularStack{ID: 1, Name: "new-name"},
		},
		{
			name:      "successful migrate without name",
			params:    map[string]any{"id": float64(1), "environmentId": float64(2), "targetEnvironmentId": float64(3)},
			mockStack: models.RegularStack{ID: 1, Name: "original"},
		},
		{
			name:        "missing id",
			params:      map[string]any{"environmentId": float64(2), "targetEnvironmentId": float64(3)},
			expectError: true,
		},
		{
			name:        "missing environmentId",
			params:      map[string]any{"id": float64(1), "targetEnvironmentId": float64(3)},
			expectError: true,
		},
		{
			name:        "missing targetEnvironmentId",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2)},
			expectError: true,
		},
		{
			name:        "invalid id",
			params:      map[string]any{"id": float64(0), "environmentId": float64(2), "targetEnvironmentId": float64(3)},
			expectError: true,
		},
		{
			name:        "invalid environmentId",
			params:      map[string]any{"id": float64(1), "environmentId": float64(-1), "targetEnvironmentId": float64(3)},
			expectError: true,
		},
		{
			name:        "invalid targetEnvironmentId",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2), "targetEnvironmentId": float64(0)},
			expectError: true,
		},
		{
			name:        "invalid name type triggers GetString error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2), "targetEnvironmentId": float64(3), "name": float64(42)},
			expectError: true,
		},
		{
			name:        "api error",
			params:      map[string]any{"id": float64(1), "environmentId": float64(2), "targetEnvironmentId": float64(3)},
			mockError:   fmt.Errorf("migration failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			idVal, hasID := tt.params["id"]
			envVal, hasEnv := tt.params["environmentId"]
			targetVal, hasTarget := tt.params["targetEnvironmentId"]
			_, nameInvalid := tt.params["name"].(float64)
			if hasID && hasEnv && hasTarget && idVal.(float64) > 0 && envVal.(float64) > 0 && targetVal.(float64) > 0 && !nameInvalid {
				name, _ := tt.params["name"].(string)
				mockClient.On("MigrateStack", int(idVal.(float64)), int(envVal.(float64)), int(targetVal.(float64)), name).Return(tt.mockStack, tt.mockError)
			}

			s := &PortainerMCPServer{cli: mockClient}
			handler := s.HandleMigrateStack()
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params
			result, err := handler(context.Background(), req)

			assert.NoError(t, err)
			if tt.expectError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleCreateRegularStack verifies the HandleCreateRegularStack MCP tool handler.
func TestHandleCreateRegularStack(t *testing.T) {
	mockStack := models.RegularStack{ID: 1, Name: "test-stack", EndpointID: 3}

	tests := []struct {
		name        string
		params      map[string]any
		mockStack   models.RegularStack
		mockError   error
		expectError bool
	}{
		{
			name: "successful creation with env",
			params: map[string]any{
				"name":          "test-stack",
				"file":          "services:\n  web:\n    image: nginx",
				"environmentId": float64(3),
				"env":           []any{map[string]any{"name": "FOO", "value": "bar"}},
			},
			mockStack: mockStack,
		},
		{
			name: "successful creation without env",
			params: map[string]any{
				"name":          "test-stack",
				"file":          "services:\n  web:\n    image: nginx",
				"environmentId": float64(3),
			},
			mockStack: mockStack,
		},
		{
			name: "missing name",
			params: map[string]any{
				"file":          "services:\n  web:\n    image: nginx",
				"environmentId": float64(3),
			},
			expectError: true,
		},
		{
			name: "missing file",
			params: map[string]any{
				"name":          "test-stack",
				"environmentId": float64(3),
			},
			expectError: true,
		},
		{
			name: "missing environmentId",
			params: map[string]any{
				"name": "test-stack",
				"file": "services:\n  web:\n    image: nginx",
			},
			expectError: true,
		},
		{
			name: "invalid environmentId",
			params: map[string]any{
				"name":          "test-stack",
				"file":          "services:\n  web:\n    image: nginx",
				"environmentId": float64(0),
			},
			expectError: true,
		},
		{
			name: "invalid env entry missing name",
			params: map[string]any{
				"name":          "test-stack",
				"file":          "services:\n  web:\n    image: nginx",
				"environmentId": float64(3),
				"env":           []any{map[string]any{"value": "bar"}},
			},
			expectError: true,
		},
		{
			name: "api error",
			params: map[string]any{
				"name":          "test-stack",
				"file":          "services:\n  web:\n    image: nginx",
				"environmentId": float64(3),
			},
			mockError:   fmt.Errorf("create failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			name, hasName := tt.params["name"].(string)
			file, hasFile := tt.params["file"].(string)
			envID, hasEnvID := tt.params["environmentId"].(float64)
			rawEnv, hasEnvParam := tt.params["env"].([]any)
			validEnv := true
			if hasEnvParam {
				for _, e := range rawEnv {
					if m, ok := e.(map[string]any); !ok || m["name"] == nil {
						validEnv = false
					}
				}
			}
			if hasName && name != "" && hasFile && file != "" && hasEnvID && envID > 0 && validEnv {
				mockClient.On("CreateRegularStack", name, file, int(envID), mock.Anything).Return(tt.mockStack, tt.mockError)
			}

			s := &PortainerMCPServer{cli: mockClient}
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params
			result, err := s.HandleCreateRegularStack()(context.Background(), req)

			assert.NoError(t, err)
			if tt.expectError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

// TestHandleUpdateRegularStack verifies the HandleUpdateRegularStack MCP tool handler.
func TestHandleUpdateRegularStack(t *testing.T) {
	mockStack := models.RegularStack{ID: 74, Name: "drafting-table", EndpointID: 3}

	tests := []struct {
		name        string
		params      map[string]any
		mockOpts    models.UpdateRegularStackOptions
		mockStack   models.RegularStack
		mockError   error
		expectError bool
	}{
		{
			name: "successful update with pullImage and prune",
			params: map[string]any{
				"id":            float64(74),
				"environmentId": float64(3),
				"file":          "services:\n  web:\n    image: nginx",
				"pullImage":     true,
				"prune":         true,
			},
			mockOpts: models.UpdateRegularStackOptions{
				EndpointID: 3, StackFileContent: "services:\n  web:\n    image: nginx",
				Env: []models.StackEnvVar{}, Prune: true, PullImage: true,
			},
			mockStack: mockStack,
		},
		{
			name: "successful update with env, no pull/prune",
			params: map[string]any{
				"id":            float64(74),
				"environmentId": float64(3),
				"file":          "services:\n  web:\n    image: nginx",
				"env":           []any{map[string]any{"name": "FOO", "value": "bar"}},
			},
			mockOpts: models.UpdateRegularStackOptions{
				EndpointID: 3, StackFileContent: "services:\n  web:\n    image: nginx",
				Env: []models.StackEnvVar{{Name: "FOO", Value: "bar"}},
			},
			mockStack: mockStack,
		},
		{
			name:        "missing id",
			params:      map[string]any{"environmentId": float64(3), "file": "services:\n  web:\n    image: nginx"},
			expectError: true,
		},
		{
			name:        "invalid id",
			params:      map[string]any{"id": float64(0), "environmentId": float64(3), "file": "services:\n  web:\n    image: nginx"},
			expectError: true,
		},
		{
			name:        "missing environmentId",
			params:      map[string]any{"id": float64(74), "file": "services:\n  web:\n    image: nginx"},
			expectError: true,
		},
		{
			name:        "invalid environmentId",
			params:      map[string]any{"id": float64(74), "environmentId": float64(0), "file": "services:\n  web:\n    image: nginx"},
			expectError: true,
		},
		{
			name:        "missing file",
			params:      map[string]any{"id": float64(74), "environmentId": float64(3)},
			expectError: true,
		},
		{
			name:        "invalid compose yaml",
			params:      map[string]any{"id": float64(74), "environmentId": float64(3), "file": "not: valid: yaml: ["},
			expectError: true,
		},
		{
			name: "api error (e.g. stack not found)",
			params: map[string]any{
				"id":            float64(74),
				"environmentId": float64(3),
				"file":          "services:\n  web:\n    image: nginx",
			},
			mockOpts: models.UpdateRegularStackOptions{
				EndpointID: 3, StackFileContent: "services:\n  web:\n    image: nginx", Env: []models.StackEnvVar{},
			},
			mockError:   fmt.Errorf("stack not found"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockPortainerClient{}
			id, hasID := tt.params["id"].(float64)
			envID, hasEnvID := tt.params["environmentId"].(float64)
			file, hasFile := tt.params["file"].(string)
			if hasID && id > 0 && hasEnvID && envID > 0 && hasFile && file != "" && file != "not: valid: yaml: [" {
				mockClient.On("UpdateRegularStack", int(id), tt.mockOpts).Return(tt.mockStack, tt.mockError)
			}

			s := &PortainerMCPServer{cli: mockClient}
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params
			result, err := s.HandleUpdateRegularStack()(context.Background(), req)

			assert.NoError(t, err)
			if tt.expectError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
			}
			mockClient.AssertExpectations(t)
		})
	}
}
