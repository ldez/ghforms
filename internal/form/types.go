package form

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// MarkdownType a Markdown text that is displayed in the form to provide extra context to the user,
	// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema#markdown
	MarkdownType = "markdown"

	// TextareaType  multi-line text field.
	// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema#textarea
	TextareaType = "textarea"

	// InputType a single-line text field.
	// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema#input
	InputType = "input"

	// DropdownType a dropdown menu.
	// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema#dropdown
	DropdownType = "dropdown"

	// CheckboxType a set of checkboxes.
	// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema#checkboxes
	CheckboxType = "checkbox"

	// UploadType a file upload field.
	// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema#upload
	UploadType = "upload"
)

const optionNone = "None"

// Form represents a GitHub issue form definition.
// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-issue-forms
type Form struct {
	Slug     string `yaml:"-"`
	Filename string `yaml:"-"`

	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Body        []*Element    `yaml:"body"`
	Assignees   StringOrSlice `yaml:"assignees,omitempty"`
	Labels      StringOrSlice `yaml:"labels,omitempty"`
	Title       string        `yaml:"title,omitempty"`
	Type        string        `yaml:"type,omitempty"`
	Projects    StringOrSlice `yaml:"projects,omitempty"`
}

// Element represents a single body element of any kind.
// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema
type Element struct {
	Type        string      `yaml:"type"`
	ID          string      `yaml:"id,omitempty"`
	Attributes  Attributes  `yaml:"attributes,omitempty"`
	Validations Validations `yaml:"validations,omitempty"`
}

// Attributes holds element attributes.
// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema
type Attributes struct {
	Label       string             `yaml:"label,omitempty"`
	Description string             `yaml:"description,omitempty"`
	Placeholder string             `yaml:"placeholder,omitempty"`
	Value       string             `yaml:"value,omitempty"`
	Render      string             `yaml:"render,omitempty"`
	Multiple    bool               `yaml:"multiple,omitempty"`
	Options     []CheckboxOrString `yaml:"options,omitempty"`
	Default     int                `yaml:"default,omitempty"`
}

// Validations holds element validation rules.
// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema
type Validations struct {
	Required bool   `yaml:"required,omitempty"`
	Accept   string `yaml:"accept,omitempty"`
}

// CheckboxOrString carries:
//   - either a plain string (dropdown option)
//   - or a checkbox option object (label + required).
type CheckboxOrString struct {
	Label    string
	Required bool
	Hidden   bool
}

// UnmarshalYAML supports both shapes for option entries:
//   - dropdown style: "Some option"
//   - checkbox style: { label: "...", required: ... }
func (o *CheckboxOrString) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var str string

		err := value.Decode(&str)
		if err != nil {
			return err
		}

		o.Label = str

		return nil

	case yaml.MappingNode:
		var raw struct {
			Label    string `yaml:"label"`
			Required bool   `yaml:"required"`
		}

		err := value.Decode(&raw)
		if err != nil {
			return err
		}

		o.Label = raw.Label
		o.Required = raw.Required

		return nil

	default:
		return fmt.Errorf("expected scalar or mapping for option at line %d", value.Line)
	}
}

// StringOrSlice array or comma-delimited string.
type StringOrSlice []string

func (s *StringOrSlice) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var str string

		err := value.Decode(&str)
		if err != nil {
			return err
		}

		if strings.TrimSpace(str) != "" {
			*s = strings.Split(str, ",")
		} else {
			*s = []string{str}
		}

		return nil

	case yaml.SequenceNode:
		var list []string

		err := value.Decode(&list)
		if err != nil {
			return err
		}

		*s = list

		return nil

	default:
		return fmt.Errorf("expected string or sequence at line %d", value.Line)
	}
}
