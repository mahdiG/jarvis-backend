package agent

// Message represents a single message in a chat conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response is returned after processing a user message through the agent.
type Response struct {
	Reply   string        `json:"reply"`
	Actions []ToolAction  `json:"actions,omitempty"`
}

// ToolAction records an action the agent performed (e.g. created a task).
type ToolAction struct {
	Tool   string `json:"tool"`
	Result string `json:"result"`
	Detail any    `json:"detail,omitempty"`
}