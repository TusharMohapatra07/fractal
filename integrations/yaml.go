package integrations

import (
	"errors"
	"io/ioutil"
	"reflect"

	"github.com/SkySingh04/fractal/interfaces"
	"github.com/SkySingh04/fractal/logger"
	"github.com/SkySingh04/fractal/registry"
	"gopkg.in/yaml.v3"
)

type YAMLSource struct {
	FilePath string `json:"yaml_source_file_path"`
}

type YAMLDestination struct {
	FilePath string `json:"yaml_output_file_path"`
}

// FetchData retrieves and processes YAML source data
func (y YAMLSource) FetchData(req interfaces.Request) (interface{}, error) {
	if req.YAMLSourceFilePath == "" {
		return nil, errors.New("missing YAML source file path")
	}

	logger.Infof("Fetching data from YAML source: %s", req.YAMLSourceFilePath)

	// Read the YAML file
	data, err := ioutil.ReadFile(req.YAMLSourceFilePath)
	if err != nil {
		return nil, err
	}

	// Validate and sanitize YAML data
	validatedData, err := ValidateYAMLData(data)
	if err != nil {
		logger.Fatalf("Validation error: %v", err)
		return nil, err
	}

	// Transform YAML data
	transformedData, err := transformYAMLData(validatedData)
	if err != nil {
		logger.Fatalf("Transformation error: %v", err)
		return nil, err
	}

	return transformedData, nil
}

// SendData writes data to a YAML destination file
func (y YAMLDestination) SendData(data interface{}, req interfaces.Request) error {
	if req.YAMLDestinationFilePath == "" {
		return errors.New("missing YAML destination file path")
	}

	logger.Infof("Sending data to YAML destination: %s", req.YAMLDestinationFilePath)

	// Write the data to the YAML file
	err := writeYAMLFile(req.YAMLDestinationFilePath, data)
	if err != nil {
		logger.Fatalf("Error writing data to YAML file: %v", err)
		return err
	}

	logger.Infof("Data successfully written to %s", req.YAMLDestinationFilePath)
	return nil
}

// ValidateYAMLData validates, sanitizes, and unmarshals YAML data
func ValidateYAMLData(data []byte) (interface{}, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, errors.New("invalid YAML format")
	}

	// Sanitize the YAML data
	sanitizedData := sanitizeYAMLData(yamlData)
	logger.Infof("Validation and sanitization successful for YAML data")
	return sanitizedData, nil
}

// sanitizeYAMLData recursively sanitizes the YAML data to ensure consistency
func sanitizeYAMLData(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// Recursively sanitize each key-value pair in the map
		for key, value := range v {
			v[key] = sanitizeYAMLData(value)
		}
		return v
	case []interface{}:
		// Recursively sanitize each element in the array
		for i, value := range v {
			v[i] = sanitizeYAMLData(value)
		}
		return v
	case string:
		// Optionally trim strings or apply further sanitization
		return v
	case float64, bool, nil:
		// Leave primitive types as-is
		return v
	default:
		// Convert unsupported types to their string representations
		logger.Warnf("Unsupported data type %T sanitized to string: %v", v, v)
		return reflect.TypeOf(v).String()
	}
}

// writeYAMLFile writes the provided data to a YAML file with proper formatting
func writeYAMLFile(filename string, data interface{}) error {
	outputData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, outputData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// transformYAMLData applies transformations to the YAML data
func transformYAMLData(data interface{}) (interface{}, error) {
	// Example transformation: Add a key-value pair if the data is a map
	if yamlMap, ok := data.(map[string]interface{}); ok {
		yamlMap["transformed"] = true
		return yamlMap, nil
	}

	// If no transformation is required, return data as is
	logger.Infof("No transformation applied to YAML data")
	return data, nil
}

func init() {
	registry.RegisterSource("YAML", YAMLSource{})
	registry.RegisterDestination("YAML", YAMLDestination{})
}
