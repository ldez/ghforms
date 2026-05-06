package form

import (
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// semanticCheck performs validation rules that the JSON schema cannot express.
func semanticCheck(source string, f *Form) error {
	unique := make(map[string]struct{})

	for i, element := range f.Body {
		id := extractID(element, i)

		if element.Type != MarkdownType {
			// Body must have unique ids.
			// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/common-validation-errors-when-creating-issue-forms#body-must-have-unique-ids
			// Body must have unique labels.
			// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/common-validation-errors-when-creating-issue-forms#body-must-have-unique-labels
			// Labels are too similar.
			// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/common-validation-errors-when-creating-issue-forms#body-must-have-unique-labels
			if _, ok := unique[id]; ok {
				return fmt.Errorf("%s: duplicate ID %q", source, id)
			}

			unique[id] = struct{}{}
		}

		switch element.Type {
		case DropdownType:
			err := validateDropdown(id, element, source)
			if err != nil {
				return err
			}

		case UploadType:
			err := validateUpload(id, element, source)
			if err != nil {
				return err
			}
		}
	}

	if len(unique) == 0 {
		// Body must contain at least one non-markdown field.
		// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/common-validation-errors-when-creating-issue-forms#body-must-contain-at-least-one-non-markdown-field
		return fmt.Errorf("%s: body must contain at least one non-markdown field", source)
	}

	return nil
}

func validateDropdown(id string, element *Element, source string) error {
	count := len(element.Attributes.Options)

	if element.Attributes.Default >= count {
		return fmt.Errorf(
			"%s: dropdown %q: default index %d exceeds the number of options (%d); valid range is 0..%d",
			source, id, element.Attributes.Default, count, count-1,
		)
	}

	// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/common-validation-errors-when-creating-issue-forms#example-of-bodyi-options-must-not-include-the-reserved-word-none-error
	if slices.ContainsFunc(element.Attributes.Options, func(el CheckboxOrString) bool {
		return strings.EqualFold(el.Label, optionNone)
	}) {
		return fmt.Errorf("%s: dropdown %q: 'None' is a reserved option name", source, id)
	}

	return nil
}

func validateUpload(id string, element *Element, source string) error {
	if element.Validations.Accept == "" {
		return nil
	}

	for accept := range strings.SplitSeq(element.Validations.Accept, ",") {
		if !strings.HasPrefix(strings.TrimSpace(accept), ".") {
			return fmt.Errorf("%s: upload %q: 'validations.accept' must be a list of file extensions",
				source, id)
		}
	}

	return nil
}

// semanticWarnings returns non-fatal issues detected in a form.
// These are configurations that GitHub accepts, but that are usually mistakes.
func semanticWarnings(source string, f *Form) {
	if f.Type != "" {
		// https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/managing-issue-types-in-an-organization
		slog.Warn("The 'type' field only works for issue forms inside an organization.",
			slog.String("actor", "forms"),
			slog.String("source", source),
		)
	}

	for i, element := range f.Body {
		switch element.Type {
		case DropdownType:
			if element.Validations.Required && !element.Attributes.Multiple {
				id := extractID(element, i)

				slog.Warn("'validations.required' is true, but 'attributes.multiple' is false; "+
					"a single-select dropdown always has a value selected, so this constraint has no effect.",
					slog.String("actor", "forms"),
					slog.String("id", id),
					slog.String("source", source),
				)
			}
		default:
			// no warnings
		}
	}
}

func extractID(element *Element, index int) string {
	id := element.ID
	if id == "" {
		id = parameterize(element.Attributes.Label)
	}

	if id == "" {
		id = fmt.Sprintf("body[%d]", index)
	}

	return id
}

// parameterize converts a string into a URL-friendly slug.
// Inspire by https://www.rubydoc.info/docs/rails/ActiveSupport%2FInflector:parameterize
func parameterize(s string) string {
	// Replace accented chars with their ASCII equivalents.
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	result, _, err := transform.String(t, s)
	if err != nil {
		slog.Error("Error",
			slog.String("actor", "forms"),
			slog.Any("error", err),
		)
	}

	const separator = "-"

	// Remove unwanted characters.
	result = regexp.MustCompile(`[^a-zA-Z0-9\-_]+`).ReplaceAllString(result, separator)

	// Deduplicate characters.
	result = regexp.MustCompile(`-{2,}`).ReplaceAllString(result, separator)

	// Remove leading/trailing separator.
	result = regexp.MustCompile(`(?i)^-|-$`).ReplaceAllString(result, "")

	return strings.ToLower(result)
}
