package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"jarvis/models"
	"jarvis/repositories"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	einoagent "github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
)

// Tool names used by the agent.
const (
	ToolCreateTask = "create_task"
)

// AgentMessage represents a single message in the chat conversation.
type AgentMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AgentResponse is the response returned after processing a user message.
type AgentResponse struct {
	Reply   string      `json:"reply"`
	Actions []ActionLog `json:"actions,omitempty"`
}

// ActionLog records an action the agent performed (e.g. created a task).
type ActionLog struct {
	Tool   string `json:"tool"`
	Result string `json:"result"`
	Detail any    `json:"detail,omitempty"`
}

// Agent is the AI chat agent powered by an LLM with tool-calling capabilities.
type Agent struct {
	chatModel model.BaseChatModel
}

// NewAgent creates a new Agent from environment variables.
//
// Required env vars:
//   - LLM_API_KEY      – API key for the OpenAI-compatible provider
//   - LLM_MODEL        – model name (e.g. "gpt-4o", "claude-sonnet-4-20250514")
//
// Optional env vars:
//   - LLM_BASE_URL     – base URL (defaults to OpenAI's API)
//   - LLM_MAX_TOKENS   – max completion tokens (default 4096)
//   - LLM_TEMPERATURE  – temperature (default 0.7)
func NewAgent() (*Agent, error) {
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		return nil, errors.New("LLM_API_KEY environment variable is required")
	}

	modelName := os.Getenv("LLM_MODEL")
	if modelName == "" {
		return nil, errors.New("LLM_MODEL environment variable is required")
	}

	baseURL := os.Getenv("LLM_BASE_URL")

	config := &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   modelName,
		BaseURL: baseURL,
		// Disable thinking/reasoning mode for providers that support it (e.g. DeepSeek).
		ExtraFields: map[string]any{
			"thinking": map[string]any{
				"type": "disabled",
			},
		},
	}

	if maxTokensStr := os.Getenv("LLM_MAX_TOKENS"); maxTokensStr != "" {
		var maxTokens int
		if err := json.Unmarshal([]byte(maxTokensStr), &maxTokens); err == nil {
			config.MaxTokens = &maxTokens
		}
	}

	if tempStr := os.Getenv("LLM_TEMPERATURE"); tempStr != "" {
		var temp float32
		if err := json.Unmarshal([]byte(tempStr), &temp); err == nil {
			config.Temperature = &temp
		}
	}

	cm, err := openai.NewChatModel(context.Background(), config)
	if err != nil {
		return nil, err
	}

	// Wrap the chat model with tool definitions so the LLM can call tools.
	chatModel, err := einoagent.ChatModelWithTools(
		cm,
		cm,
		[]*schema.ToolInfo{createTaskToolInfo()},
	)
	if err != nil {
		return nil, err
	}

	return &Agent{chatModel: chatModel}, nil
}

// Chat processes a user message and returns the agent's response.
// It sends the conversation history to the LLM, handles any tool calls
// the model makes, and returns the final reply along with action logs.
func (a *Agent) Chat(ctx context.Context, history []AgentMessage) (*AgentResponse, error) {
	messages := convertHistory(history)

	resp, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}

	actions := make([]ActionLog, 0)

	// The agent may return a text reply.
	reply := resp.Content

	// Process any tool calls the model made.
	for _, tc := range resp.ToolCalls {
		result, err := a.executeTool(ctx, tc)
		if err != nil {
			slog.Error("tool execution failed", "tool", tc.Function.Name, "error", err)
			result = err.Error()
		}

		// Feed the tool result back to the model so it can produce a final reply.
		// Preserve reasoning content if the model returned any (e.g. DeepSeek thinking mode).
		assistantMsg := &schema.Message{
			Role:             schema.Assistant,
			Content:          "",
			ToolCalls:        []schema.ToolCall{tc},
			ReasoningContent: resp.ReasoningContent,
		}
		messages = append(messages,
			assistantMsg,
			schema.ToolMessage(result, tc.ID),
		)

		action := ActionLog{
			Tool:   tc.Function.Name,
			Result: result,
		}

		// Attempt to parse the tool call arguments for detail.
		var args any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err == nil {
			action.Detail = args
		}

		actions = append(actions, action)
	}

	// If we ran tools, get the final response from the model.
	if len(resp.ToolCalls) > 0 {
		finalResp, err := a.chatModel.Generate(ctx, messages)
		if err != nil {
			return nil, err
		}
		reply = finalResp.Content
	}

	return &AgentResponse{
		Reply:   reply,
		Actions: actions,
	}, nil
}

// executeTool dispatches a tool call to the appropriate handler.
func (a *Agent) executeTool(ctx context.Context, tc schema.ToolCall) (string, error) {
	switch tc.Function.Name {
	case ToolCreateTask:
		return executeCreateTask(ctx, tc.Function.Arguments)
	default:
		return "", errors.New("unknown tool: " + tc.Function.Name)
	}
}

// createTaskArgs is the JSON schema for the create_task tool.
type createTaskArgs struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// executeCreateTask creates a task via the repository and returns the result.
func executeCreateTask(_ context.Context, arguments string) (string, error) {
	var args createTaskArgs
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", err
	}

	if args.Title == "" {
		return "", errors.New("title is required")
	}

	task := models.Task{
		Title:       args.Title,
		Description: args.Description,
	}

	created, err := repositories.CreateTask(task)
	if err != nil {
		return "", err
	}

	result, _ := json.Marshal(created)
	return string(result), nil
}

// createTaskToolInfo returns the ToolInfo for the create_task tool.
func createTaskToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: ToolCreateTask,
		Desc: "Create a new task in the task management system. " +
			"Use this when the user asks to create, add, or make a task.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"title": {
				Type:     schema.String,
				Desc:     "The title of the task",
				Required: true,
			},
			"description": {
				Type: schema.String,
				Desc: "A detailed description of the task",
			},
		}),
	}
}

// convertHistory converts []AgentMessage to []*schema.Message for the LLM.
func convertHistory(history []AgentMessage) []*schema.Message {
	if len(history) == 0 {
		return []*schema.Message{
			schema.SystemMessage("You are an AI assistant for a task management system called Jarvis. " +
				"You can help users manage their tasks. When a user asks to create a task, " +
				"use the create_task tool to create it."),
		}
	}

	messages := make([]*schema.Message, 0, len(history)+1)
	// Always prepend a system prompt.
	messages = append(messages, schema.SystemMessage(
		"You are an AI assistant for a task management system called Jarvis. "+
			"You can help users manage their tasks. When a user asks to create a task, "+
			"use the create_task tool to create it."))

	for _, msg := range history {
		switch msg.Role {
		case "user":
			messages = append(messages, schema.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, schema.AssistantMessage(msg.Content, nil))
		case "system":
			messages = append(messages, schema.SystemMessage(msg.Content))
		}
	}

	return messages
}
