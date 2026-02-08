package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/tidwall/gjson"
)

// ToolAdapter adapts an MCP tool to Eino's InvokableTool interface
type ToolAdapter struct {
	mcpTool    Tool
	client     *Client
	serverName string
}

// NewToolAdapter creates a new tool adapter for an MCP tool
func NewToolAdapter(mcpTool Tool, client *Client, serverName string) *ToolAdapter {
	return &ToolAdapter{
		mcpTool:    mcpTool,
		client:     client,
		serverName: serverName,
	}
}

// Info implements tool.InvokableTool interface
// Returns the tool information in Eino's expected format
func (a *ToolAdapter) Info(ctx context.Context) (*schema.ToolInfo, error) {
	// Build unique name by prefixing with server name
	prefixedName := fmt.Sprintf("mcp_%s_%s", a.serverName, a.mcpTool.Name)

	// Convert JSON Schema to Eino ParameterInfo
	params, err := convertJSONSchemaToEinoParams(a.mcpTool.InputSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to convert parameters: %w", err)
	}

	return &schema.ToolInfo{
		Name:        prefixedName,
		Desc:        fmt.Sprintf("[MCP:%s] %s", a.serverName, a.mcpTool.Description),
		ParamsOneOf: schema.NewParamsOneOfByParams(params),
	}, nil
}

// InvokableRun implements tool.InvokableTool interface
// Executes the MCP tool call
func (a *ToolAdapter) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// Parse the arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Call the MCP tool
	result, err := a.client.CallTool(ctx, a.mcpTool.Name, args)
	if err != nil {
		return "", fmt.Errorf("MCP tool call failed: %w", err)
	}

	// Format the result as a string for the AI
	return formatMCPResult(result, a.mcpTool.Name), nil
}

// formatMCPResult formats the MCP tool result as a readable string
func formatMCPResult(result interface{}, toolName string) string {
	// Handle nil result
	if result == nil {
		return fmt.Sprintf("Tool %s executed successfully with no result", toolName)
	}

	// Check if result is a string
	if str, ok := result.(string); ok {
		return fmt.Sprintf("Tool %s result: %s", toolName, str)
	}

	// Check if result is a map/object
	if obj, ok := result.(map[string]interface{}); ok {
		// Try to format as JSON
		jsonBytes, err := json.MarshalIndent(obj, "", "  ")
		if err == nil {
			return fmt.Sprintf("Tool %s result:\n```json\n%s\n```", toolName, string(jsonBytes))
		}
		// Fallback to string representation
		return fmt.Sprintf("Tool %s result: %+v", toolName, obj)
	}

	// Check if result is an array
	if arr, ok := result.([]interface{}); ok {
		jsonBytes, err := json.MarshalIndent(arr, "", "  ")
		if err == nil {
			return fmt.Sprintf("Tool %s result:\n```json\n%s\n```", toolName, string(jsonBytes))
		}
		return fmt.Sprintf("Tool %s returned %d items", toolName, len(arr))
	}

	// Default: convert to JSON string
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err == nil {
		return fmt.Sprintf("Tool %s result:\n```json\n%s\n```", toolName, string(jsonBytes))
	}

	return fmt.Sprintf("Tool %s result: %+v", toolName, result)
}

// convertJSONSchemaToEinoParams converts MCP JSON Schema to Eino ParameterInfo
// This handles the conversion from JSON Schema to Eino's expected parameter format
func convertJSONSchemaToEinoParams(inputSchema interface{}) (map[string]*schema.ParameterInfo, error) {
	params := make(map[string]*schema.ParameterInfo)

	if inputSchema == nil {
		return params, nil
	}

	// Handle as JSON string (some MCP servers send schema as string)
	if schemaStr, ok := inputSchema.(string); ok {
		var schemaObj map[string]interface{}
		if err := json.Unmarshal([]byte(schemaStr), &schemaObj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON schema string: %w", err)
		}
		inputSchema = schemaObj
	}

	// Expect JSON Schema with "properties" and "required" fields
	schemaObj, ok := inputSchema.(map[string]interface{})
	if !ok {
		return params, nil
	}

	// Get properties
	properties, ok := schemaObj["properties"].(map[string]*schema.ParameterInfo)
	if !ok {
		return properties, nil // Empty or invalid schema
	}

	// Get required fields
	requiredFields := make(map[string]bool)
	if required, ok := schemaObj["required"].([]interface{}); ok {
		for _, field := range required {
			if fieldName, ok := field.(string); ok {
				requiredFields[fieldName] = true
			}
		}
	}

	// Convert each property to ParameterInfo
	for paramName, paramDef := range properties {
		paramInfo, err := convertPropertyToParamInfo(paramName, paramDef, requiredFields[paramName])
		if err != nil {
			continue // Log or handle error
		}
		params[paramName] = paramInfo
	}

	return params, nil
}

// convertPropertyToParamInfo converts a single property definition to ParameterInfo
func convertPropertyToParamInfo(name string, def interface{}, required bool) (*schema.ParameterInfo, error) {
	paramDef, ok := def.(map[string]interface{})
	if !ok {
		return &schema.ParameterInfo{
			Type:     "string",
			Desc:     "",
			Required: false,
		}, nil
	}

	// Get type
	paramType := "string"
	if paramTypeRaw, ok := paramDef["type"].(string); ok {
		paramType = paramTypeRaw
	}

	// Handle array types
	if paramType == "array" {
		paramType = "array"
	}

	// Get description
	description := ""
	if desc, ok := paramDef["description"].(string); ok {
		description = desc
	}

	// Check if there's an enum (allowed values)
	if enum, ok := paramDef["enum"].([]interface{}); ok && len(enum) > 0 {
		enumValues := ""
		for i, val := range enum {
			if i > 0 {
				enumValues += ", "
			}
			if strVal, ok := val.(string); ok {
				enumValues += strVal
			}
		}
		description = fmt.Sprintf("%s (Allowed values: %s)", description, enumValues)
	}

	return &schema.ParameterInfo{
		Type:     schema.DataType(paramType),
		Desc:     description,
		Required: required,
	}, nil
}

// CreateToolAdapters creates Eino tools from MCP client
func CreateToolAdapters(client *Client, serverName string) []tool.InvokableTool {
	mcpTools := client.GetTools()
	adapters := make([]tool.InvokableTool, 0, len(mcpTools))

	for _, mcpTool := range mcpTools {
		adapter := NewToolAdapter(mcpTool, client, serverName)
		adapters = append(adapters, adapter)
	}

	return adapters
}

// FindToolInJSON extracts tool name and arguments from a JSON-RPC tool call
// This helper is used when parsing AI model responses
func FindToolInJSON(jsonStr string) (toolName string, args map[string]interface{}, err error) {
	parsed := gjson.Get(jsonStr, "tool_calls.0.function.name")
	if !parsed.Exists() {
		return "", nil, fmt.Errorf("no tool call found")
	}

	toolName = parsed.String()

	argsParsed := gjson.Get(jsonStr, "tool_calls.0.function.arguments")
	if argsParsed.Exists() {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(argsParsed.String()), &args); err != nil {
			return toolName, nil, err
		}
		return toolName, args, nil
	}

	return toolName, nil, nil
}
