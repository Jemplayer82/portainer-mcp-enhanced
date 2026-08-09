package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/jmrplens/portainer-mcp-enhanced/pkg/portainer/models"
	"github.com/jmrplens/portainer-mcp-enhanced/pkg/toolgen"
)

// AddStackFeatures registers the stack management tools on the MCP server.
func (s *PortainerMCPServer) AddStackFeatures() {
	s.addToolIfExists(ToolListStacks, s.HandleGetStacks())
	s.addToolIfExists(ToolListRegularStacks, s.HandleListRegularStacks())
	s.addToolIfExists(ToolGetStackFile, s.HandleGetStackFile())
	s.addToolIfExists(ToolGetStack, s.HandleInspectStack())
	s.addToolIfExists(ToolInspectStackFile, s.HandleInspectStackFile())

	if !s.readOnly {
		s.addToolIfExists(ToolCreateStack, s.HandleCreateStack())
		s.addToolIfExists(ToolUpdateStack, s.HandleUpdateStack())
		s.addToolIfExists(ToolDeleteStack, s.HandleDeleteStack())
		s.addToolIfExists(ToolUpdateStackGit, s.HandleUpdateStackGit())
		s.addToolIfExists(ToolRedeployStackGit, s.HandleRedeployStackGit())
		s.addToolIfExists(ToolStartStack, s.HandleStartStack())
		s.addToolIfExists(ToolStopStack, s.HandleStopStack())
		s.addToolIfExists(ToolMigrateStack, s.HandleMigrateStack())
		s.addToolIfExists(ToolCreateRegularStack, s.HandleCreateRegularStack())
	}
}

// HandleGetStacks returns an MCP tool handler that retrieves stacks.
func (s *PortainerMCPServer) HandleGetStacks() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stacks, err := s.cli.GetStacks()
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to get stacks", err), nil
		}

		return jsonResult(stacks, "failed to marshal stacks")
	}
}

// HandleListRegularStacks returns an MCP tool handler that lists regular stacks.
func (s *PortainerMCPServer) HandleListRegularStacks() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stacks, err := s.cli.GetRegularStacks()
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to list regular stacks", err), nil
		}

		return jsonResult(stacks, "failed to marshal regular stacks")
	}
}

// HandleGetStackFile returns an MCP tool handler that retrieves stack file.
func (s *PortainerMCPServer) HandleGetStackFile() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		id, err := parser.GetInt("id", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid id parameter", err), nil
		}
		if err := validatePositiveID("id", id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		stackFile, err := s.cli.GetStackFile(id)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to get stack file", err), nil
		}

		return mcp.NewToolResultText(stackFile), nil
	}
}

// HandleCreateStack returns an MCP tool handler that creates stack.
func (s *PortainerMCPServer) HandleCreateStack() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		name, err := parser.GetString("name", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid name parameter", err), nil
		}
		if err := validateName(name); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		file, err := parser.GetString("file", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid file parameter", err), nil
		}
		if err := validateComposeYAML(file); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		environmentGroupIds, err := parser.GetArrayOfIntegers("environmentGroupIds", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid environmentGroupIds parameter", err), nil
		}

		id, err := s.cli.CreateStack(name, file, environmentGroupIds)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("error creating stack", err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Stack created successfully with ID: %d", id)), nil
	}
}

// HandleUpdateStack returns an MCP tool handler that updates stack.
func (s *PortainerMCPServer) HandleUpdateStack() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		id, err := parser.GetInt("id", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid id parameter", err), nil
		}
		if err := validatePositiveID("id", id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		file, err := parser.GetString("file", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid file parameter", err), nil
		}
		if err := validateComposeYAML(file); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		environmentGroupIds, err := parser.GetArrayOfIntegers("environmentGroupIds", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid environmentGroupIds parameter", err), nil
		}

		err = s.cli.UpdateStack(id, file, environmentGroupIds)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to update stack", err), nil
		}

		return mcp.NewToolResultText("Stack updated successfully"), nil
	}
}

// HandleInspectStack returns an MCP tool handler that retrieves detailed information about stack.
func (s *PortainerMCPServer) HandleInspectStack() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		id, err := parser.GetInt("id", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid id parameter", err), nil
		}
		if err := validatePositiveID("id", id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		stack, err := s.cli.InspectStack(id)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to inspect stack", err), nil
		}

		return jsonResult(stack, "failed to marshal stack")
	}
}

// HandleDeleteStack returns an MCP tool handler that deletes stack.
func (s *PortainerMCPServer) HandleDeleteStack() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		id, err := parser.GetInt("id", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid id parameter", err), nil
		}
		if err := validatePositiveID("id", id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		endpointID, err := parser.GetInt("environmentId", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid environmentId parameter", err), nil
		}
		if err := validatePositiveID("environmentId", endpointID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		removeVolumes, err := parser.GetBoolean("removeVolumes", false)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid removeVolumes parameter", err), nil
		}

		err = s.cli.DeleteStack(id, models.DeleteStackOptions{
			EndpointID:    endpointID,
			RemoveVolumes: removeVolumes,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to delete stack", err), nil
		}

		return mcp.NewToolResultText("Stack deleted successfully"), nil
	}
}

// HandleInspectStackFile returns an MCP tool handler that retrieves detailed information about stack file.
func (s *PortainerMCPServer) HandleInspectStackFile() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		id, err := parser.GetInt("id", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid id parameter", err), nil
		}
		if err := validatePositiveID("id", id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		content, err := s.cli.InspectStackFile(id)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to inspect stack file", err), nil
		}

		return mcp.NewToolResultText(content), nil
	}
}

// HandleUpdateStackGit returns an MCP tool handler that updates stack git.
func (s *PortainerMCPServer) HandleUpdateStackGit() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		id, err := parser.GetInt("id", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid id parameter", err), nil
		}
		if err := validatePositiveID("id", id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		endpointID, err := parser.GetInt("environmentId", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid environmentId parameter", err), nil
		}
		if err := validatePositiveID("environmentId", endpointID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		referenceName, err := parser.GetString("referenceName", false)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid referenceName parameter", err), nil
		}

		prune, err := parser.GetBoolean("prune", false)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid prune parameter", err), nil
		}

		stack, err := s.cli.UpdateStackGit(id, models.UpdateStackGitOptions{
			EndpointID:    endpointID,
			ReferenceName: referenceName,
			Prune:         prune,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to update stack git", err), nil
		}

		return jsonResult(stack, "failed to marshal stack")
	}
}

// redeployStackGitResponse wraps a redeployed stack with the outcome of any
// force-pull attempts performed ahead of the redeploy. models.RegularStack is
// embedded anonymously so its fields are promoted to the top level of the
// marshaled JSON, keeping the response shape unchanged for callers that don't
// care about pull results.
type redeployStackGitResponse struct {
	models.RegularStack
	ImagePulls []imagePullResult `json:"imagePulls,omitempty"`
}

// imagePullResult reports the outcome of force-pulling a single stack image
// ahead of a git redeploy.
type imagePullResult struct {
	Image   string `json:"image"`
	Pulled  bool   `json:"pulled"`
	Message string `json:"message"`
}

// HandleRedeployStackGit returns an MCP tool handler that redeploys stack git.
//
// Portainer's git-redeploy "pullImage" option has been observed to silently
// keep serving a stale cached image for mutable tags (e.g. ":latest") even
// when it reports a successful, healthy redeploy with a fresh container ID
// and the correct git commit hash. When pullImage is requested, this handler
// closes that gap by force-pulling every mutable-tag image the stack's
// compose file references via the raw Docker Engine API before asking
// Portainer to redeploy, and reports what actually happened in the response
// instead of only trusting Portainer's flag.
func (s *PortainerMCPServer) HandleRedeployStackGit() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		id, err := parser.GetInt("id", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid id parameter", err), nil
		}
		if err := validatePositiveID("id", id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		endpointID, err := parser.GetInt("environmentId", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid environmentId parameter", err), nil
		}
		if err := validatePositiveID("environmentId", endpointID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		pullImage, err := parser.GetBoolean("pullImage", false)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid pullImage parameter", err), nil
		}

		prune, err := parser.GetBoolean("prune", false)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid prune parameter", err), nil
		}

		var imagePulls []imagePullResult
		if pullImage {
			composeContent, fileErr := s.cli.GetStackFile(id)
			if fileErr != nil {
				log.Warn().Err(fileErr).Int("id", id).
					Msg("could not read stack file to force-pull images before git redeploy; falling back to Portainer's built-in pullImage")
			} else {
				imagePulls = s.forcePullStackImages(endpointID, composeContent)
			}
		}

		stack, err := s.cli.RedeployStackGit(id, models.RedeployStackGitOptions{
			EndpointID: endpointID,
			PullImage:  pullImage,
			Prune:      prune,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to redeploy stack", err), nil
		}

		if imagePulls == nil {
			return jsonResult(stack, "failed to marshal stack")
		}

		return jsonResult(redeployStackGitResponse{
			RegularStack: stack,
			ImagePulls:   imagePulls,
		}, "failed to marshal stack")
	}
}

// composeFile is the minimal shape needed to extract image references from a
// docker-compose file for force-pull verification.
type composeFile struct {
	Services map[string]struct {
		Image string `yaml:"image"`
	} `yaml:"services"`
}

// extractComposeImages parses a compose YAML file and returns the unique,
// mutable-tag image references declared by its services (service.image
// fields). Build-only services (no "image" key) are skipped. Images pinned
// by digest (repo@sha256:...) are excluded since they are already
// content-addressed and cannot go stale. Invalid YAML yields an empty slice.
func extractComposeImages(content string) []string {
	var parsed composeFile
	if err := yaml.Unmarshal([]byte(content), &parsed); err != nil {
		return nil
	}

	seen := make(map[string]bool, len(parsed.Services))
	images := make([]string, 0, len(parsed.Services))
	for _, svc := range parsed.Services {
		image := strings.TrimSpace(svc.Image)
		if image == "" || strings.Contains(image, "@sha256:") || seen[image] {
			continue
		}
		seen[image] = true
		images = append(images, image)
	}
	return images
}

// splitImageRef splits a "repo[:tag]" image reference into repository and
// tag, defaulting the tag to "latest" when omitted. A colon after the last
// "/" is treated as the tag separator; a colon before it (e.g.
// "myregistry:5000/repo") is part of a registry host:port and left alone.
// Digest references (repo@sha256:...) are not handled here -- callers should
// exclude those before calling.
func splitImageRef(image string) (repo, tag string) {
	lastSlash := strings.LastIndex(image, "/")
	lastColon := strings.LastIndex(image, ":")
	if lastColon > lastSlash {
		return image[:lastColon], image[lastColon+1:]
	}
	return image, "latest"
}

// forcePullStackImages force-pulls every mutable-tag image referenced by a
// stack's compose file via the Docker Engine API, ahead of a git redeploy.
// See HandleRedeployStackGit for why this exists. Pull failures are not
// fatal and do not stop the caller from proceeding to redeploy -- a
// private-registry image may legitimately fail here since this raw proxy
// call carries no registry auth, whereas Portainer's own compose-based pull
// uses its stored credentials. Failures are reported per-image and logged.
func (s *PortainerMCPServer) forcePullStackImages(endpointID int, composeContent string) []imagePullResult {
	images := extractComposeImages(composeContent)
	results := make([]imagePullResult, 0, len(images))

	for _, image := range images {
		repo, tag := splitImageRef(image)

		response, err := s.cli.ProxyDockerRequest(models.DockerProxyRequestOptions{
			EnvironmentID: endpointID,
			Method:        http.MethodPost,
			Path:          "/images/create",
			QueryParams:   map[string]string{"fromImage": repo, "tag": tag},
		})
		if err != nil {
			log.Warn().Err(err).Str("image", image).Msg("force-pull before git redeploy failed")
			results = append(results, imagePullResult{Image: image, Pulled: false, Message: err.Error()})
			continue
		}

		status, pullErr := readImageCreateStatus(response)
		results = append(results, imagePullResult{Image: image, Pulled: pullErr == nil, Message: status})
		if pullErr != nil {
			log.Warn().Err(pullErr).Str("image", image).Msg("force-pull before git redeploy failed")
		}
	}

	return results
}

// readImageCreateStatus reads a POST /images/create response body -- a
// stream of newline-delimited JSON status/progress events -- and returns the
// last status message. It returns a non-nil error if the response carries a
// non-2xx status or any event includes a Docker "error" field.
func readImageCreateStatus(response *http.Response) (string, error) {
	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxProxyResponseSize+1))
	if err != nil {
		return "", fmt.Errorf("failed to read image pull response: %w", err)
	}

	var lastStatus string
	var pullErr error
	for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var event struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		}
		if jsonErr := json.Unmarshal([]byte(line), &event); jsonErr != nil {
			continue
		}
		if event.Error != "" {
			pullErr = fmt.Errorf("%s", event.Error)
		}
		if event.Status != "" {
			lastStatus = event.Status
		}
	}

	if response.StatusCode >= 400 && pullErr == nil {
		pullErr = fmt.Errorf("image pull request failed with status %s", response.Status)
	}
	if pullErr != nil {
		return pullErr.Error(), pullErr
	}
	if lastStatus == "" {
		return "", fmt.Errorf("no status reported by Docker daemon")
	}
	return lastStatus, nil
}

// HandleStartStack returns an MCP tool handler that starts stack.
func (s *PortainerMCPServer) HandleStartStack() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		id, err := parser.GetInt("id", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid id parameter", err), nil
		}
		if err := validatePositiveID("id", id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		endpointID, err := parser.GetInt("environmentId", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid environmentId parameter", err), nil
		}
		if err := validatePositiveID("environmentId", endpointID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		stack, err := s.cli.StartStack(id, endpointID)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to start stack", err), nil
		}

		return jsonResult(stack, "failed to marshal stack")
	}
}

// HandleStopStack returns an MCP tool handler that stops stack.
func (s *PortainerMCPServer) HandleStopStack() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		id, err := parser.GetInt("id", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid id parameter", err), nil
		}
		if err := validatePositiveID("id", id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		endpointID, err := parser.GetInt("environmentId", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid environmentId parameter", err), nil
		}
		if err := validatePositiveID("environmentId", endpointID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		stack, err := s.cli.StopStack(id, endpointID)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to stop stack", err), nil
		}

		return jsonResult(stack, "failed to marshal stack")
	}
}

// HandleMigrateStack returns an MCP tool handler that migrates stack.
func (s *PortainerMCPServer) HandleMigrateStack() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		id, err := parser.GetInt("id", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid id parameter", err), nil
		}
		if err := validatePositiveID("id", id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		endpointID, err := parser.GetInt("environmentId", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid environmentId parameter", err), nil
		}
		if err := validatePositiveID("environmentId", endpointID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		targetEndpointID, err := parser.GetInt("targetEnvironmentId", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid targetEnvironmentId parameter", err), nil
		}
		if err := validatePositiveID("targetEnvironmentId", targetEndpointID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		name, err := parser.GetString("name", false)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid name parameter", err), nil
		}

		stack, err := s.cli.MigrateStack(id, endpointID, targetEndpointID, name)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to migrate stack", err), nil
		}

		return jsonResult(stack, "failed to marshal stack")
	}
}

// parseStackEnvVars converts an array of {"name", "value"} objects (as decoded from
// JSON) into a slice of models.StackEnvVar. An empty/nil input yields an empty slice.
func parseStackEnvVars(items []any) ([]models.StackEnvVar, error) {
	env := make([]models.StackEnvVar, 0, len(items))

	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid env entry: %v", item)
		}

		name, ok := obj["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("env entry missing required 'name' string field: %v", item)
		}

		value, _ := obj["value"].(string)

		env = append(env, models.StackEnvVar{Name: name, Value: value})
	}

	return env, nil
}

// HandleCreateRegularStack returns an MCP tool handler that creates a standalone
// (non-edge) stack on a specific environment.
func (s *PortainerMCPServer) HandleCreateRegularStack() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parser := toolgen.NewParameterParser(request)

		name, err := parser.GetString("name", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid name parameter", err), nil
		}
		if err := validateName(name); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		file, err := parser.GetString("file", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid file parameter", err), nil
		}
		if err := validateComposeYAML(file); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		endpointID, err := parser.GetInt("environmentId", true)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid environmentId parameter", err), nil
		}
		if err := validatePositiveID("environmentId", endpointID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		rawEnv, err := parser.GetArrayOfObjects("env", false)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid env parameter", err), nil
		}
		env, err := parseStackEnvVars(rawEnv)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("invalid env parameter", err), nil
		}

		stack, err := s.cli.CreateRegularStack(name, file, endpointID, env)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to create regular stack", err), nil
		}

		return jsonResult(stack, "failed to marshal stack")
	}
}
