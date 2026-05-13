package agent

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"jarvis/models"
	"jarvis/repositories"
	"jarvis/utils"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

// Tool names exposed to the LLM.
const (
	toolNameCreateTask = "create_task"
	toolNameGetTasks   = "get_tasks"
	toolNameGetTask    = "get_task"
	toolNameUpdateTask = "update_task"
	toolNameDeleteTask = "delete_task"
)

// toolSchemaFromModel generates an Eino ParamsOneOf by reflecting a model struct
// and optionally stripping fields inherited from models.Base (ID, timestamps, soft-delete).
// When shouldExcludeBase is true, auto-generated fields are removed from the schema
// so the LLM does not provide them. When false, all fields (including ID) are exposed.
func getToolSchemaFromModel[Model any](shouldExcludeBase bool) *schema.ParamsOneOf {
	var model Model
	modelType := reflect.TypeOf(model)

	basePropertyNames := getBaseFieldPropertyNames()

	reflector := &jsonschema.Reflector{
		// Expand the root struct so the schema has type:"object" at the root.
		// Without this, the library returns a $ref to the definition, which has
		// no Type field — causing OpenAI to reject it with "type: null".
		ExpandedStruct: true,
	}

	if shouldExcludeBase {
		// At the root level, remove fields that come from Base since they are
		// auto-generated or ignored at creation time.
		reflector.SchemaModifier = func(
			jsonFieldName string,
			structType reflect.Type,
			structTag reflect.StructTag,
			propertySchema *jsonschema.Schema,
		) {
			if jsonFieldName != "_root" {
				return
			}
			for _, baseFieldName := range basePropertyNames {
				propertySchema.Properties.Delete(baseFieldName)
			}
		}
	}

	modelSchema := reflector.ReflectFromType(modelType)

	return schema.NewParamsOneOfByJSONSchema(modelSchema)
}

// baseFieldPropertyNames returns the Go field names of models.Base using reflect.
// The jsonschema library uses f.Name (Go field name) when no json tag is
// present, so this gives us the exact property names to strip from the tool
// schema.
func getBaseFieldPropertyNames() []string {
	baseType := reflect.TypeOf(models.Base{})
	fieldNames := make([]string, baseType.NumField())
	for i := range baseType.NumField() {
		fieldNames[i] = baseType.Field(i).Name
	}
	return fieldNames
}

// getToolInfoCreateTask returns the ToolInfo for the create_task tool.
func getToolInfoCreateTask() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: toolNameCreateTask,
		Desc: "Create a new task in the task management system. " +
			"Use this when the user asks to create, add, or make a task or todo",
		ParamsOneOf: getToolSchemaFromModel[models.Task](true),
	}
}

// getToolInfoGetTasks returns the ToolInfo for the get_tasks tool.
func getToolInfoGetTasks() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: toolNameGetTasks,
		Desc: "Get tasks. Use limit=0 to return all tasks, or limit=N to return the first N tasks. " +
			"Use offset for pagination (e.g. offset=3 skips the first 3 tasks).",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"limit": {
				Desc:     "Maximum number of tasks to return. 0 returns all tasks.",
				Type:     schema.Integer,
				Required: true,
			},
			"offset": {
				Desc: "Number of tasks to skip before returning results. 0 or omitted starts from the beginning.",
				Type: schema.Integer,
			},
		}),
	}
}

// getToolInfoGetTask returns the ToolInfo for the get_task tool.
func getToolInfoGetTask() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name:        toolNameGetTask,
		Desc:        "Get a single task by its ID. Use this when the user asks to get, fetch, or retrieve a specific task",
		ParamsOneOf: getToolSchemaFromModel[models.Task](false),
	}
}

// getToolInfoUpdateTask returns the ToolInfo for the update_task tool.
func getToolInfoUpdateTask() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name:        toolNameUpdateTask,
		Desc:        "Update an existing task. Use this when the user asks to update, edit, or modify a task",
		ParamsOneOf: getToolSchemaFromModel[models.Task](false),
	}
}

// getToolInfoDeleteTask returns the ToolInfo for the delete_task tool.
func getToolInfoDeleteTask() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name:        toolNameDeleteTask,
		Desc:        "Delete a task by its ID. Use this when the user asks to delete, remove, or archive a task",
		ParamsOneOf: getToolSchemaFromModel[models.Task](false),
	}
}

// executeToolCreateTask parses the tool arguments, creates a task via the
// repository, and returns the JSON-encoded result.
func executeToolCreateTask(_ context.Context, argumentsJSON string) (string, error) {
	var task models.Task
	err := json.Unmarshal([]byte(argumentsJSON), &task)
	if err != nil {
		return "", utils.WrapError(err)
	}

	if task.Title == "" {
		return "", utils.WrapError(errors.New("title is required"))
	}

	created, err := repositories.CreateTask(task)
	if err != nil {
		return "", utils.WrapError(err)
	}

	result, marshalError := json.Marshal(created)
	if marshalError != nil {
		return "", utils.WrapError(marshalError)
	}
	return string(result), nil
}

// executeToolGetTasks returns tasks as a JSON-encoded array, with optional
// limit/offset for pagination.
func executeToolGetTasks(_ context.Context, argumentsJSON string) (string, error) {
	var params struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}
	if argumentsJSON != "" {
		if err := json.Unmarshal([]byte(argumentsJSON), &params); err != nil {
			return "", utils.WrapError(err)
		}
	}

	tasks, err := repositories.GetTasks(params.Limit, params.Offset)
	if err != nil {
		return "", utils.WrapError(err)
	}

	result, marshalError := json.Marshal(tasks)
	if marshalError != nil {
		return "", utils.WrapError(marshalError)
	}
	return string(result), nil
}

// executeToolGetTask parses the tool arguments, retrieves a single task by ID,
// and returns the JSON-encoded result.
func executeToolGetTask(_ context.Context, argumentsJSON string) (string, error) {
	var condition models.Task
	err := json.Unmarshal([]byte(argumentsJSON), &condition)
	if err != nil {
		return "", utils.WrapError(err)
	}

	if condition.ID == "" {
		return "", utils.WrapError(errors.New("task ID is required"))
	}

	task, err := repositories.GetTask(condition)
	if err != nil {
		return "", utils.WrapError(err)
	}

	result, marshalError := json.Marshal(task)
	if marshalError != nil {
		return "", utils.WrapError(marshalError)
	}
	return string(result), nil
}

// executeToolUpdateTask parses the tool arguments, updates a task via the
// repository, and returns the JSON-encoded result.
func executeToolUpdateTask(_ context.Context, argumentsJSON string) (string, error) {
	var task models.Task
	err := json.Unmarshal([]byte(argumentsJSON), &task)
	if err != nil {
		return "", utils.WrapError(err)
	}

	if task.ID == "" {
		return "", utils.WrapError(errors.New("task ID is required for update"))
	}

	updated, err := repositories.UpdateTask(task)
	if err != nil {
		return "", utils.WrapError(err)
	}

	result, marshalError := json.Marshal(updated)
	if marshalError != nil {
		return "", utils.WrapError(marshalError)
	}
	return string(result), nil
}

// executeToolDeleteTask parses the tool arguments, deletes a task by ID,
// and returns a success message.
func executeToolDeleteTask(_ context.Context, argumentsJSON string) (string, error) {
	var condition models.Task
	err := json.Unmarshal([]byte(argumentsJSON), &condition)
	if err != nil {
		return "", utils.WrapError(err)
	}

	if condition.ID == "" {
		return "", utils.WrapError(errors.New("task ID is required for deletion"))
	}

	err = repositories.DeleteTask(condition)
	if err != nil {
		return "", utils.WrapError(err)
	}

	return "task deleted successfully", nil
}