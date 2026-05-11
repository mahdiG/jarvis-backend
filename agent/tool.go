package agent

import (
	"context"
	"encoding/json"
	"errors"

	"jarvis/models"
	"jarvis/repositories"

	"github.com/cloudwego/eino/schema"
)

// Tool names exposed to the LLM.
const (
	toolNameCreateTask = "create_task"
)

// createTaskArguments is the JSON schema for the create_task tool.
type createTaskArguments struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// toolCreateTask returns the ToolInfo for the create_task tool.
func toolCreateTask() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: toolNameCreateTask,
		Desc: "Create a new task in the task management system. " +
			"Use this when the user asks to create, add, or make a task or todo",
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

// executeToolCreateTask parses the tool arguments, creates a task via the
// repository, and returns the JSON-encoded result.
func executeToolCreateTask(_ context.Context, argumentsJSON string) (string, error) {
	var arguments createTaskArguments
	if err := json.Unmarshal([]byte(argumentsJSON), &arguments); err != nil {
		return "", err
	}

	if arguments.Title == "" {
		return "", errors.New("title is required")
	}

	task := models.Task{
		Title:       arguments.Title,
		Description: arguments.Description,
	}

	created, err := repositories.CreateTask(task)
	if err != nil {
		return "", err
	}

	result, _ := json.Marshal(created)
	return string(result), nil
}
