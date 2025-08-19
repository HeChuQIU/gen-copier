package copier_generate

import (
	"fmt"
	"go/ast"
	"log"
	"strings"
)

type FieldInfo struct {
	Name string
	Type string
	Tag  string
}

func getFieldInfo(field *ast.Field) FieldInfo {
	info := FieldInfo{}

	// 获取字段名
	if len(field.Names) > 0 {
		info.Name = field.Names[0].Name
	}

	// 获取字段类型
	if expr, ok := field.Type.(*ast.Ident); ok {
		info.Type = expr.Name
	} else if expr, ok := field.Type.(*ast.SelectorExpr); ok {
		info.Type = fmt.Sprintf("%s.%s", expr.X, expr.Sel.Name)
	} else if expr, ok := field.Type.(*ast.ArrayType); ok {
		info.Type = "[]" + expr.Elt.(*ast.Ident).Name
	}

	// 获取标签
	if field.Tag != nil {
		tag := strings.Trim(field.Tag.Value, "`")
		info.Tag = tag
	}

	return info
}

func GenerateCopier(sourceTypeName string, sourceType *ast.StructType,
	targetTypeName string, targetType *ast.StructType) (string, error) {
	log.SetPrefix("[generator]")
	log.Printf("Generating copier between:\n  Source: %s\n  Target: %s\n",
		sourceTypeName, targetTypeName)

	var builder strings.Builder
	copyMap := make(map[string]string)
	targetFields := make(map[string]FieldInfo)

	// 首先收集目标类型的所有字段信息
	for _, field := range targetType.Fields.List {
		info := getFieldInfo(field)
		if info.Name != "" {
			targetFields[info.Name] = info
		}
	}

	// 处理有标签的字段
	log.Println("Processing tagged fields...")
	for _, field := range sourceType.Fields.List {
		info := getFieldInfo(field)
		if info.Name == "" {
			continue
		}

		log.Printf("  Source field: %s (type: %s)\n", info.Name, info.Type)

		if tagValue := extractTagValue(info.Tag, "gen-copier"); tagValue != "" {
			log.Printf("    Found tag: %s -> %s\n", info.Name, tagValue)
			// 检查目标字段是否存在
			targetInfo, found := targetFields[tagValue]
			if !found {
				return "", fmt.Errorf("target field %s not found in type %s", tagValue, targetTypeName)
			}
			// 检查类型是否匹配
			if info.Type != targetInfo.Type {
				return "", fmt.Errorf("type mismatch: source field %s (%s) and target field %s (%s) have different types",
					info.Name, info.Type, tagValue, targetInfo.Type)
			}
			copyMap[info.Name] = tagValue
			log.Printf("    Added mapping: %s -> %s\n", info.Name, tagValue)
		}
	}

	// 处理无标签的同名字段
	log.Println("Processing untagged fields with matching names...")
	for _, field := range sourceType.Fields.List {
		info := getFieldInfo(field)
		if info.Name == "" {
			continue
		}

		if _, ok := copyMap[info.Name]; ok {
			continue // 跳过已在copyMap中的字段
		}

		// 检查目标类型是否有同名字段
		if targetInfo, found := targetFields[info.Name]; found {
			log.Printf("  Found matching field: %s (type: %s)\n", info.Name, info.Type)
			// 检查类型是否匹配
			if info.Type == targetInfo.Type {
				copyMap[info.Name] = info.Name
				log.Printf("    Added mapping: %s -> %s\n", info.Name, info.Name)
			} else {
				log.Printf("    Type mismatch: %s (%s) vs %s (%s)\n",
					info.Name, info.Type, targetInfo.Name, targetInfo.Type)
			}
		}
	}

	log.Printf("Final field mappings: %v\n", copyMap)

	// 修改点1：简化源类型名称，去掉包名前缀
	sourceTypeShort := strings.TrimPrefix(sourceTypeName, "packageA.")

	// 生成代码
	builder.WriteString(fmt.Sprintf("func (source *%s) CopyTo(target *%s) {\n",
		sourceTypeShort, targetTypeName))

	for sourceField, targetField := range copyMap {
		builder.WriteString(fmt.Sprintf("    target.%s = source.%s\n", targetField, sourceField))
	}

	builder.WriteString("}\n")

	log.Println("Copier function generated successfully")
	return builder.String(), nil
}

func extractTagValue(tag, key string) string {
	if tag == "" {
		return ""
	}

	// 简单实现标签解析，实际项目中可能需要更完整的解析
	parts := strings.Split(tag, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, key+":") {
			return strings.Trim(strings.TrimPrefix(part, key+":"), `"`)
		}
	}
	return ""
}
