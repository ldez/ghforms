package form

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type Loader struct {
	issueValidator      *SchemaValidator
	discussionValidator *SchemaValidator
	chooserValidator    *SchemaValidator
}

func New() (*Loader, error) {
	issueValidator, err := NewSchemaIssueValidator()
	if err != nil {
		return nil, err
	}

	discussionValidator, err := NewSchemaDiscussionValidator()
	if err != nil {
		return nil, err
	}

	chooserValidator, err := NewSchemaChooserValidator()
	if err != nil {
		return nil, err
	}

	return &Loader{
		issueValidator:      issueValidator,
		discussionValidator: discussionValidator,
		chooserValidator:    chooserValidator,
	}, nil
}

// Load reads all issue forms from a directory, validates them against the
// embedded JSON schema, and returns them sorted by file name.
//
// Files named `config.yml` and `config.yaml` are skipped: they are issue
// chooser configurations, not forms.
func (l *Loader) Load(dir string) ([]*Form, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	var forms []*Form

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}

		isDiscussion := strings.Contains(dir, "DISCUSSION_TEMPLATE")

		path := filepath.Join(dir, name)

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		if !isDiscussion && strings.EqualFold(strings.TrimSuffix(name, ext), "config") {
			err = l.chooserValidator.validate(path, data)
			if err != nil {
				return nil, err
			}

			continue
		}

		form, err := l.validateForm(isDiscussion, name, path, data)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}

	slices.SortFunc(forms, func(a, b *Form) int {
		return strings.Compare(a.Filename, b.Filename)
	})

	return forms, nil
}

func (l *Loader) validateForm(isDiscussion bool, name, path string, data []byte) (*Form, error) {
	err := l.getValidator(isDiscussion).validate(path, data)
	if err != nil {
		return nil, err
	}

	form := new(Form)

	err = yaml.Unmarshal(data, form)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	form.Filename = name
	form.Slug = strings.TrimSuffix(name, filepath.Ext(name))

	if isDiscussion {
		form.Name = "💬 " + strings.TrimSuffix(name, filepath.Ext(name))
		form.Description = "DISCUSSION_TEMPLATE"
	}

	err = semanticCheck(path, form)
	if err != nil {
		return nil, err
	}

	semanticWarnings(path, form)

	addExtra(form)

	return form, nil
}

func (l *Loader) getValidator(isDiscussion bool) *SchemaValidator {
	if isDiscussion {
		return l.discussionValidator
	}

	return l.issueValidator
}

func addExtra(form *Form) {
	for _, element := range form.Body {
		if element.Type == DropdownType && !element.Validations.Required && !element.Attributes.Multiple {
			element.Attributes.Options = slices.Insert(element.Attributes.Options, 0, CheckboxOrString{
				Label: optionNone,
			})
		}
	}
}
