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
)

// toolSchemaFromModel generates an Eino ParamsOneOf by reflecting a model struct
// and stripping fields inherited from models.Base (ID, timestamps, soft-delete).
// This lets us reuse the model type as the tool argument schema without exposing
// auto-generated fields to the LLM.
func getToolSchemaFromModel[Model any]() *schema.ParamsOneOf {
	var model Model
	modelType := reflect.TypeOf(model)

	basePropertyNames := getBaseFieldPropertyNames()

	reflector := &jsonschema.Reflector{
		// Expand the root struct so the schema has type:"object" at the root.
		// Without this, the library returns a $ref to the definition, which has
		// no Type field — causing OpenAI to reject it with "type: null".
		ExpandedStruct: true,
	}

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
		ParamsOneOf: getToolSchemaFromModel[models.Task](),
	}
}

// executeToolCreateTask parses the tool arguments, creates a task via the
// repository, and returns the JSON-encoded result.
func executeToolCreateTask(_ context.Context, argumentsJSON string) (string, error) {
	var task models.Task
	if err := json.Unmarshal([]byte(argumentsJSON), &task); err != nil {
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
