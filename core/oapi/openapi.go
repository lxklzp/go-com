package oapi

import (
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
)

var Openapi openapiHandler

type openapiHandler struct{}

const (
	componentsSchemasPrefix = "#/components/schemas/"
	refKey                  = `"$ref"`
	OpenapiVersion          = "3.0.3"
	MaxSchemaDeep           = 4 // schema最大引用深度
	SecurityBearer          = `,"components":{"securitySchemes":{"bearerAuth":{"type":"http","scheme":"bearer","bearerFormat":"token"}}},"security":[{"bearerAuth":[]}]}`
)

// ComponentsSchemasRefToInline 将ComponentsSchemas的引用转换成内联模式
func (handler openapiHandler) ComponentsSchemasRefToInline(components *openapi3.Components) error {
	if components == nil {
		return nil
	}
	for i := 0; i < MaxSchemaDeep; i++ {
		if err := handler.componentsSchemasRefToInlineOnce(components); err != nil {
			return err
		}
		bytes, err := components.MarshalJSON()
		if err != nil {
			return err
		}
		if !strings.Contains(string(bytes), refKey) {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("ComponentsSchema 引用深度超过%d层，解析失败。", MaxSchemaDeep))
}

// 将ComponentsSchemas的引用转换成内联模式
func (handler openapiHandler) componentsSchemasRefToInlineOnce(components *openapi3.Components) error {
	var err error
	for k := range components.Schemas {
		if err = handler.schemaRefToInlineOnce(components.Schemas[k], k, components); err != nil {
			return err
		}
	}
	return nil
}

// 将Schema的引用转换成内联模式
func (handler openapiHandler) schemaRefToInlineOnce(schemaRef *openapi3.SchemaRef, schemaName string, components *openapi3.Components) error {
	if schemaRef == nil {
		return nil
	}

	// 检测schema最大引用深度
	if schemaName != "" && !handler.checkSchemaDeep(schemaRef, components, 1) {
		return errors.New(fmt.Sprintf("ComponentsSchema[%s] 引用深度超过%d层，解析失败。", schemaName, MaxSchemaDeep))
	}

	if schemaRef.Ref != "" {
		name := strings.TrimPrefix(schemaRef.Ref, componentsSchemasPrefix)
		schemaRef.Value = components.Schemas[name].Value
		schemaRef.Ref = ""
		return nil
	}
	schema := schemaRef.Value
	for _, property := range schema.Properties {
		if property.Ref != "" {
			name := strings.TrimPrefix(property.Ref, componentsSchemasPrefix)
			property.Value = components.Schemas[name].Value
			property.Ref = ""
		} else if property.Value.Items != nil && property.Value.Items.Ref != "" {
			name := strings.TrimPrefix(property.Value.Items.Ref, componentsSchemasPrefix)
			property.Value.Items.Value = components.Schemas[name].Value
			property.Value.Items.Ref = ""
		} else if property.Value.AdditionalProperties.Schema != nil {
			items := property.Value.AdditionalProperties.Schema.Value.Items
			if items != nil && items.Ref != "" {
				name := strings.TrimPrefix(items.Ref, componentsSchemasPrefix)
				items.Value = components.Schemas[name].Value
				items.Ref = ""
			}
		}
	}
	return nil
}

// 检测schema最大引用深度
func (handler openapiHandler) checkSchemaDeep(schemaRef *openapi3.SchemaRef, components *openapi3.Components, deep int) bool {
	if deep > MaxSchemaDeep {
		return false
	}

	var pass bool
	if schemaRef.Ref != "" {
		name := strings.TrimPrefix(schemaRef.Ref, componentsSchemasPrefix)
		if pass = handler.checkSchemaDeep(components.Schemas[name], components, deep+1); !pass {
			return false
		}
		return true
	}
	schema := schemaRef.Value
	for _, property := range schema.Properties {
		if property.Ref != "" {
			name := strings.TrimPrefix(property.Ref, componentsSchemasPrefix)
			if pass = handler.checkSchemaDeep(components.Schemas[name], components, deep+1); !pass {
				return false
			}
		} else if property.Value.Items != nil && property.Value.Items.Ref != "" {
			name := strings.TrimPrefix(property.Value.Items.Ref, componentsSchemasPrefix)
			if pass = handler.checkSchemaDeep(components.Schemas[name], components, deep+1); !pass {
				return false
			}
		}
	}
	return true
}

// PathsSchemaRefToInline 将PathsSchema的引用转换成内联模式
func (handler openapiHandler) PathsSchemaRefToInline(paths *openapi3.Paths, components *openapi3.Components) {
	if paths == nil || components == nil {
		return
	}

	for _, path := range paths.Map() {
		handler.operationToInlineOnce(path.Get, components)
		handler.operationToInlineOnce(path.Post, components)
		handler.operationToInlineOnce(path.Put, components)
		handler.operationToInlineOnce(path.Delete, components)
	}
}

// 将operation的引用转换成内联模式
func (handler openapiHandler) operationToInlineOnce(operation *openapi3.Operation, components *openapi3.Components) {
	if operation == nil {
		return
	}
	for k := range operation.Parameters {
		handler.schemaRefToInlineOnce(operation.Parameters[k].Value.Schema, "", components)
	}
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		for k := range operation.RequestBody.Value.Content {
			handler.schemaRefToInlineOnce(operation.RequestBody.Value.Content[k].Schema, "", components)
		}
	}
	for k := range operation.Responses.Map() {
		if operation.Responses.Value(k).Value != nil {
			for m := range operation.Responses.Value(k).Value.Content {
				handler.schemaRefToInlineOnce(operation.Responses.Value(k).Value.Content.Get(m).Schema, "", components)
			}
		}
	}
}
