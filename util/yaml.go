package util

import (
	"bytes"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

// GenerateYAMLTemplate 生成 YAML 模板并为 omitempty 字段增加注释
func GenerateYAMLTemplate(v interface{}) string {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Struct {
		panic("input must be a struct")
	}

	// 获取字段的 omitempty 信息
	omitemptyFields := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		if strings.Contains(tag, "omitempty") {
			yamlKey := strings.Split(tag, ",")[0]
			if yamlKey == "" {
				yamlKey = field.Name
			}
			omitemptyFields[yamlKey] = true
		}
	}

	// 序列化 YAML
	var yamlData bytes.Buffer
	encoder := yaml.NewEncoder(&yamlData)
	if err := encoder.Encode(v); err != nil {
		panic(err)
	}
	encoder.Close()

	// 添加注释
	var buffer bytes.Buffer
	lines := strings.Split(yamlData.String(), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "") && strings.Contains(trimmed, ":") {
			key := strings.Split(trimmed, ":")[0]
			key = strings.TrimSpace(key)
			if omitemptyFields[key] {
				line += " # 允许缺省值"
			}
		}
		buffer.WriteString(line + "\n")
	}

	return buffer.String()
}
