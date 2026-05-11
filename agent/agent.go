package agent

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"jarvis/configs"
	"jarvis/utils"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	einoagent "github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
)

// chatModel is the configured LLM chat model with tool definitions.
var chatModel model.BaseChatModel

// Init initializes the AI agent from environment configuration.
//
// Required environment variables:
//   - LLM_API_KEY – API key for the OpenAI-compatible provider.
//   - LLM_MODEL   – model name (e.g. "gpt-4o", "claude-sonnet-4-20250514").
//
// Optional environment variables:
//   - LLM_BASE_URL     – base URL (defaults to OpenAI's API).
//   - LLM_MAX_TOKENS   – max completion tokens (default 4096).
//   - LLM_TEMPERATURE  – temperature (default 0.7).
func Init() error {
	if configs.Envs.LLMApiKey == "" {
		return utils.WrapError(errors.New("LLM_API_KEY is required"))
	}
	if configs.Envs.LLMModel == "" {
		return utils.WrapError(errors.New("LLM_MODEL is required"))
	}

	openAIConfig := buildOpenAIConfig()

	openaiCompatibleChatModel, err := openai.NewChatModel(context.Background(), openAIConfig)
	if err != nil {
		return utils.WrapError(err)
	}

	// Wrap the chat model with tool definitions so the LLM can call tools.
	chatModelWithTools, err := einoagent.ChatModelWithTools(
		openaiCompatibleChatModel,
		openaiCompatibleChatModel,
		[]*schema.ToolInfo{getToolInfoCreateTask()},
	)
	if err != nil {
		return utils.WrapError(err)
	}

	chatModel = chatModelWithTools
	return nil
}

// Chat sends a message history to the LLM, processes any tool calls the model
// makes, and returns the final reply along with action logs.
// TODO: get/set history in database (pg).
func Chat(ctx context.Context, history []Message) (*Response, error) {
	messages := convertMessagesToSchemaMessages(history)

	actions := make([]ToolAction, 0)
	var reply string

	for {
		response, err := chatModel.Generate(ctx, messages)
		if err != nil {
			return nil, utils.WrapError(err)
		}

		slog.Debug("LLM response", "response", response)
		slog.Debug("LLM response.ToolCalls", "tool calls", len(response.ToolCalls))

		// No tool calls — this is the final text reply.
		if len(response.ToolCalls) == 0 {
			reply = response.Content
			break
		}

		// Process each tool call and feed results back so the model can
		// continue with follow-up tool calls (e.g. parent → child tasks).
		for _, toolCall := range response.ToolCalls {
			result, err := executeTool(ctx, toolCall)
			if err != nil {
				slog.Error("tool execution failed", "tool", toolCall.Function.Name, "error", err)
				result = err.Error()
			}

			// Feed the tool result back to the model so it can produce a final reply.
			// Preserve reasoning content if the model returned any (e.g. DeepSeek thinking mode).
			assistantMessage := &schema.Message{
				Role:             schema.Assistant,
				Content:          "",
				ToolCalls:        []schema.ToolCall{toolCall},
				ReasoningContent: response.ReasoningContent,
			}
			messages = append(messages,
				assistantMessage,
				schema.ToolMessage(result, toolCall.ID),
			)

			action := ToolAction{
				Tool:   toolCall.Function.Name,
				Result: result,
			}

			// Attempt to parse the tool call arguments for the detail field.
			var arguments any
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &arguments); err == nil {
				action.Detail = arguments
			}

			actions = append(actions, action)
		}
	}

	return &Response{
		Reply:   reply,
		Actions: actions,
	}, nil
}

// executeTool dispatches a tool call to the appropriate handler.
func executeTool(ctx context.Context, toolCall schema.ToolCall) (string, error) {
	switch toolCall.Function.Name {
	case toolNameCreateTask:
		return executeToolCreateTask(ctx, toolCall.Function.Arguments)
	default:
		return "", utils.WrapError(errors.New("unknown tool: " + toolCall.Function.Name))
	}
}

// buildOpenAIConfig creates an OpenAI-compatible chat model config from
// environment configuration.
func buildOpenAIConfig() *openai.ChatModelConfig {
	config := &openai.ChatModelConfig{
		APIKey:  configs.Envs.LLMApiKey,
		Model:   configs.Envs.LLMModel,
		BaseURL: configs.Envs.LLMBaseURL,
		// Disable thinking/reasoning mode for providers that support it (e.g. DeepSeek).
		ExtraFields: map[string]any{
			"thinking": map[string]any{
				"type": "disabled",
			},
		},
	}

	if configs.Envs.LLMMaxTokens > 0 {
		config.MaxTokens = &configs.Envs.LLMMaxTokens
	}

	if configs.Envs.LLMTemperature > 0 {
		temperature := float32(configs.Envs.LLMTemperature)
		config.Temperature = &temperature
	}

	return config
}

// convertMessagesToSchemaMessages converts a slice of Messages to Eino schema messages.
// It always prepends a system prompt before the conversation history.
func convertMessagesToSchemaMessages(history []Message) []*schema.Message {
	systemPrompt := "You are an AI assistant for a task management system called Jarvis. " +
		"You can help users manage their tasks. When a user asks to create a task, " +
		"use the create_task tool to create it."

	messages := make([]*schema.Message, 0, len(history)+1)
	messages = append(messages, schema.SystemMessage(systemPrompt))

	for _, message := range history {
		switch message.Role {
		case "user":
			messages = append(messages, schema.UserMessage(message.Content))
		case "assistant":
			messages = append(messages, schema.AssistantMessage(message.Content, nil))
		case "system":
			messages = append(messages, schema.SystemMessage(message.Content))
		}
	}

	return messages
}
