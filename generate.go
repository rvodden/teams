//go:build ignore

package main

import (
	"fmt"
	"github.com/rvodden/teams/model"
	"gopkg.in/yaml.v3"
	"log"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"text/template"
)

// generateTemplate creates a Go code template string for a given entity type.
//
// It generates a template that can be used to create a slice of entities
// in the generated_data package.
//
// Parameters:
//   - pluralName: A string representing the plural name of the entity type,
//     used as the variable name for the slice in the generated code.
//   - entity: An interface{} representing an instance of the entity type,
//     used to reflect on its structure and generate appropriate field templates.
//
// Returns:
//
//	A string containing the generated Go code template, which includes:
//	- A package declaration for generated_data
//	- An import statement for the model package
//	- A variable declaration for a slice of the entity type
//	- A template for populating the slice with entity instances
func generateTemplate(pluralName string, entity interface{}) string {
	entityType := reflect.TypeOf(entity)
	var fields []string
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		fieldName := field.Name
		fieldType := field.Type

		var fieldTemplate string
		switch fieldType.Kind() {
		case reflect.Slice:
			fieldTemplate = fmt.Sprintf(`%s: []%s{ {{- range .%s }}"{{ . }}",{{- end }} }`, fieldName, fieldType.Elem(), fieldName)
		default:
			fieldTemplate = fmt.Sprintf(`%s: "{{ .%s }}"`, fieldName, fieldName)
		}
		fields = append(fields, fieldTemplate)
	}

	return fmt.Sprintf(`package generated_data

import "github.com/rvodden/teams/model"

var %s = []%s{
{{- range . }}
    {%s},
{{- end }}
}
`, pluralName, entityType, strings.Join(fields, ", "))
}

// sanitizeEntity trims whitespace from string fields and string slice elements of the given entity.
//
// This function uses reflection to iterate over the fields of the provided entity struct.
// It trims leading and trailing whitespace from string fields and elements of string slices.
//
// Parameters:
//   - entity: A pointer to a struct of any type. The struct's fields will be sanitized.
//
// Returns:
//   - An error if the provided entity is not a pointer to a struct, or nil if the sanitization was successful.
//
// Example usage:
//
//	type Person struct {
//	    Name string
//	    Hobbies []string
//	}
//	p := &Person{Name: "  John  ", Hobbies: []string{" reading ", "  writing "}}
//	err := sanitizeEntity(p)
//	// After sanitization: p.Name == "John", p.Hobbies == []string{"reading", "writing"}
func sanitizeEntity[entityType any](entity *entityType) error {
	val := reflect.ValueOf(entity)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct, got %T", entity)
	}
	val = val.Elem() // Dereference the pointer

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		switch field.Kind() {
		case reflect.String:
			if field.CanSet() {
				field.SetString(strings.TrimSpace(field.String()))
			}
		case reflect.Slice:
			if field.Type().Elem().Kind() == reflect.String {
				for j := 0; j < field.Len(); j++ {
					strVal := field.Index(j)
					if strVal.CanSet() {
						strVal.SetString(strings.TrimSpace(strVal.String()))
					}
				}
			}
		}
	}

	return nil
}

// generateCodeString generates a code string from a slice of entities using a provided template and sanitizer function.
//
// This function takes a slice of entities, applies a sanitizer function to each entity,
// and then executes a provided template with the sanitized entities to generate a code string.
//
// Parameters:
//   - entities: A slice of entities of type entityType to be processed and used in the template.
//   - templateString: A string containing the template to be used for code generation.
//   - sanitizer: A function that takes a pointer to an entity and sanitizes it, returning an error if the sanitization fails.
//
// Returns:
//   - string: The generated code string after applying the template to the sanitized entities.
//   - error: An error if template parsing, entity sanitization, or template execution fails. Returns nil if successful.
func generateCodeString[entityType any](entities []entityType, templateString string, sanitizer func(*entityType) error) (string, error) {
	tmpl := template.Must(template.New("template").Parse(templateString))

	// Trim fields just like the generator does
	for i := range entities {
		slog.Info(fmt.Sprintf("entity %v", i), "entity", entities[i])
		err := sanitizer(&entities[i])
		if err != nil {
			log.Fatalf("failed to sanitize entity %v: %v", i, err)
		}
		slog.Info(fmt.Sprintf("sanitized entity %v", i), "entity", entities[i])
	}

	var sb strings.Builder
	err := tmpl.Execute(&sb, entities)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

// generateCodeFile generates a Go code file containing a slice of entities based on YAML data.
//
// This function reads entity data from a YAML file, unmarshals it into a slice of entities,
// generates a Go code template, and writes the resulting code to a file in the generated_data package.
//
// Parameters:
//   - name: A string representing the singular name of the entity type (currently unused in the function body).
//   - pluralName: A string representing the plural name of the entity type, used for file naming and code generation.
//   - exampleEntity: An instance of the entity type, used as a template for code generation.
//
// The function does not return any values, but it will log fatal errors if any step in the process fails.

func generateCodeFile[entityType any](name string, pluralName string, exampleEntity entityType) {
	sourceDataFile := "data/" + pluralName + ".yaml"

	data, err := os.ReadFile(sourceDataFile)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	slog.Info("data:", "data", data)

	var listOfEntities []entityType
	if err := yaml.Unmarshal(data, &listOfEntities); err != nil {
		log.Fatalf("failed to unmarshal YAML: %v", err)
	}
	slog.Info("entities:", "entities", data)

	template := generateTemplate(strings.Title(pluralName), exampleEntity)
	entityCode, err := generateCodeString(listOfEntities, template, sanitizeEntity)
	if err != nil {
		log.Fatalf("failed to generate listOfEntities code: %v", err)
	}

	destinationDataFile := "internal/generated_data/" + pluralName + "_data.go"
	f, err := os.Create(destinationDataFile)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatalf("failed to close file file: %v", err)
		}
	}(f)

	_, err = f.WriteString(entityCode)
	if err != nil {
		log.Fatalf("failed to write code to file: %v", err)
	}
}

func main() {
	generateCodeFile("person", "people", model.Person{})
	generateCodeFile("team", "teams", model.Team{})
}
