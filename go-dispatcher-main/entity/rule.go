package entity

type Field struct {
	Name  string
	Value string
}

type ForwardInput struct {
	Fields        []Field
	HasAttachment bool `json:"has_attachment"`
	Content       string
}

type ForwardOutput struct {
	CustomFieldID    string `json:"custom_field_id"`
	CustomFieldValue string `json:"custom_field_value"`
}

type Forward struct {
	Input  ForwardInput  `json:"input"`
	Output ForwardOutput `json:"output"`
}

type Rule struct {
	Name    string
	Forward Forward
}
