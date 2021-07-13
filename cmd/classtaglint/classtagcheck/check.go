package classtagcheck

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/hashicorp/eventlogger/filters/encrypt"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:             "classtaglint",
	Doc:              "analyze usage of class struct tags",
	Run:              run,
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	RunDespiteErrors: true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.StructType)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		structType := n.(*ast.StructType)

		for _, field := range structType.Fields.List {
			checkStructField(pass, field)
		}
	})

	return nil, nil
}

func checkStructField(pass *analysis.Pass, field *ast.Field) error {
	// ensure the field has a string tag
	tag := field.Tag

	if tag == nil || tag.Kind != token.STRING {
		return nil
	}

	// chop beginning and ending "`"
	tagValue := field.Tag.Value
	tagValue = strings.TrimPrefix(tagValue, "`")
	tagValue = strings.TrimSuffix(tagValue, "`")

	// look for multiple class tags
	var foundClassDeclared int

	// iterate over tag content
	tagFields := strings.Fields(tagValue)
	for _, tf := range tagFields {
		if strings.HasPrefix(tf, "class:") {
			foundClassDeclared++

			classOptionsValue, err := strconv.Unquote(strings.TrimPrefix(tf, "class:"))
			if err != nil {
				return fmt.Errorf("failed to parse class tag value while unqoute options: %w", err)
			}
			classOptions := strings.Split(classOptionsValue, ",")

			var (
				classification string
				operation      string
			)

			if len(classOptions) >= 1 {
				classification = classOptions[0]
			}

			if len(classOptions) >= 2 {
				operation = classOptions[1]
			}

			if len(classOptions) >= 3 {
				pass.Reportf(field.Tag.Pos(), "too many classification options given: %d", len(classOptions))
			}

			switch encrypt.DataClassification(classification) {
			case encrypt.UnknownClassification:
				if operation != "" {
					pass.Reportf(field.Tag.Pos(), "filter operations invalid on unknown data classifications")
				}
			case encrypt.PublicClassification:
				if operation != "" {
					pass.Reportf(field.Tag.Pos(), "filter operations invalid on public data classifications")
				}
			case encrypt.SecretClassification:
				checkStructFieldTypeFilterable(pass, field)
			case encrypt.SensitiveClassification:
				checkStructFieldTypeFilterable(pass, field)
			default:
				pass.Reportf(field.Tag.Pos(), "invalid data classification: %q", classification)
			}

			if operation != "" {
				switch encrypt.FilterOperation(operation) {
				case encrypt.RedactOperation:
					// TODO
				case encrypt.EncryptOperation:
					// TODO
				case encrypt.HmacSha256Operation:
					// TODO
				case encrypt.NoOperation:
					// TODO
				default:
					pass.Reportf(field.Tag.Pos(), "invalid filter operation: %q", operation)
				}
			}

		}
	}

	if foundClassDeclared > 1 {
		pass.Reportf(field.Tag.Pos(), "found %d data classifications for single field", foundClassDeclared)
	}

	return nil
}

func checkStructFieldTypeFilterable(pass *analysis.Pass, field *ast.Field) {
	fieldType := pass.TypesInfo.TypeOf(field.Type)
	if fieldType == nil {
		return
	}

	switch fieldType.String() {
	case "string":
	case "[]string":
	case "[]byte":
	case "[][]byte":
	default:
		pass.Reportf(field.Tag.Pos(), "invalid data classification for non-filterable type")
	}
}
