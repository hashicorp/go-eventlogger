package classtagcheck

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/hashicorp/eventlogger/filters/encrypt"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/ssa"
)

var Analyzer = &analysis.Analyzer{
	Name:             "classtaglint",
	Doc:              "analyze usage of class struct tags and taggable interfaces",
	Run:              run,
	Requires:         []*analysis.Analyzer{inspect.Analyzer, buildssa.Analyzer},
	RunDespiteErrors: true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	structFilter := []ast.Node{
		(*ast.StructType)(nil),
	}

	inspect.Preorder(structFilter, func(n ast.Node) {
		structType := n.(*ast.StructType)

		for _, field := range structType.Fields.List {
			checkStructFieldTags(pass, field)
		}
	})

	ssa := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)

	checkWithSSA(pass, ssa)

	return nil, nil
}

func checkStructFieldTags(pass *analysis.Pass, field *ast.Field) error {
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
	case "*google.golang.org/protobuf/types/known/wrapperspb.StringValue":
	case "*google.golang.org/protobuf/types/known/wrapperspb.BytesValue":
	default:
		pass.Reportf(field.Tag.Pos(), "invalid data classification for non-filterable type")
	}
}

const (
	tagsFuncString   = "func() ([]github.com/hashicorp/eventlogger/filters/encrypt.PointerTag, error)"
	tagsTypeString   = "[]github.com/hashicorp/eventlogger/filters/encrypt.PointerTag"
	pointerTagString = "*github.com/hashicorp/eventlogger/filters/encrypt.PointerTag"
	errorTypeString  = "error"
)

func checkWithSSA(pass *analysis.Pass, builtSSA *buildssa.SSA) error {
	for _, fn := range builtSSA.SrcFuncs {
		// NOTE: debug SSA
		// fn.WriteTo(os.Stderr)

		if fn.Name() == "Tags" && fmt.Sprintf("%v", fn.Signature) != tagsFuncString {
			continue // did not match
		}

		// TODO: require the function has a recv?
		if recv := fn.Signature.Recv(); recv != nil {
			switch recv.Type().(type) {
			case *types.Named:
				named := recv.Type().(*types.Named)
				switch named.Underlying().(type) {
				case *types.Map:
					// likely valid taggable type
				case *types.Struct:
					// likely valid taggable type, but need to check the definition
					for id, obj := range pass.TypesInfo.Defs {
						if named.Obj().Id() == id.Name {
							typeSpec, ok := id.Obj.Decl.(*ast.TypeSpec)
							if !ok {
								pass.Reportf(obj.Pos(), "invalid taggable reciever object declaration type %T", id.Obj.Decl)
								continue
							}
							switch fmt.Sprintf("%v", pass.TypesInfo.TypeOf(typeSpec.Type)) {
							case "google.golang.org/protobuf/types/known/structpb.Struct":
								// valid type
							default:
								pass.Reportf(obj.Pos(), "invalid taggable reciever struct declaration type %s", pass.TypesInfo.TypeOf(typeSpec.Type))

							}
						}
					}
				default:
					pass.Reportf(fn.Pos(), "invalid type %T (%T) used for taggable interface", recv.Type(), recv.Type().Underlying())
				}
			default:
				pass.Reportf(fn.Pos(), "invalid type %T (%T) used for taggable interface", recv.Type(), recv.Type().Underlying())
			}
		}

		for _, block := range fn.DomPreorder() {
			for _, instr := range block.Instrs {
				if store, ok := instr.(*ssa.Store); ok {
					if fieldAddr, ok := store.Addr.(*ssa.FieldAddr); ok {
						checkFieldAddr(pass, builtSSA, fieldAddr)
					}
				}
				if unop, ok := instr.(*ssa.UnOp); ok && unop.Op == token.MUL {
					checkFieldUnOp(pass, builtSSA, unop)
				}
			}
		}
	}

	return nil
}

func checkFieldUnOp(pass *analysis.Pass, builtSSA *buildssa.SSA, unop *ssa.UnOp) error {
	if fmt.Sprintf("%v", unop.X.Type()) == pointerTagString {
		switch unopValue := unop.X.(type) {
		case *ssa.Global:
			structRef, ok := unopValue.Type().Underlying().(*types.Pointer).Elem().Underlying().(*types.Struct)
			if ok {
				for expr := range pass.TypesInfo.Types {
					ident, ok := expr.(*ast.Ident)
					if ok {
						if ident.Name == unopValue.Name() {
							if valueSpec, ok := ident.Obj.Decl.(*ast.ValueSpec); ok {
								for _, value := range valueSpec.Values {
									if cl, ok := value.(*ast.CompositeLit); ok {
										checkCompositLiteral(pass, builtSSA, cl)
									}
								}
							}
						}
					}
				}
				for fi := 0; fi < structRef.NumFields(); fi++ {
					field := structRef.Field(fi)
					switch field.Name() {
					case "Pointer":
					case "Classification":
					case "Filter":
					}
				}
			}
		}
	}
	return nil
}

func checkFieldAddr(pass *analysis.Pass, builtSSA *buildssa.SSA, fieldAddr *ssa.FieldAddr) error {
	structRef, ok := fieldAddr.X.Type().Underlying().(*types.Pointer).Elem().Underlying().(*types.Struct)
	if !ok {
		return nil
	}

	structFieldName := structRef.Field(fieldAddr.Field).Name()

	fieldRefs := fieldAddr.Referrers()
	if fieldRefs == nil || len(*fieldRefs) == 0 {
		return nil
	}
	switch structFieldName {
	case "Pointer":
		for _, fieldRef := range *fieldRefs {
			fieldRefStore, ok := (fieldRef).(*ssa.Store)
			if !ok {
				continue
			}
			switch value := fieldRefStore.Val.(type) {
			case *ssa.Const: // static string
				switch value.Value.Kind() {
				case constant.String:
					strValue, err := strconv.Unquote(value.Value.ExactString())
					if err != nil {
						return fmt.Errorf("failed to unquote pointer field value string: %w", err)
					}
					if strValue == "" {
						pass.Reportf(fieldRef.Pos(), "empty pointerstructure pointer string")
					}
				default:
					// TODO: implement other types?
				}
			case *ssa.Call: // dynamic string
				// TODO: resolve dynamic strings or report?
			default:
				// TODO: implement other types?
			}
		}
	case "Classification":
		for _, fieldRef := range *fieldRefs {
			fieldRefStore, ok := (fieldRef).(*ssa.Store)
			if !ok {
				continue
			}
			switch value := fieldRefStore.Val.(type) {
			case *ssa.Const: // static string
				switch value.Value.Kind() {
				case constant.String:
					strValue, err := strconv.Unquote(value.Value.ExactString())
					if err != nil {
						return fmt.Errorf("failed to unquote classification field value string: %w", err)
					}
					switch encrypt.DataClassification(strValue) {
					case encrypt.PublicClassification:
						// TODO
					case encrypt.SecretClassification:
						// TODO
					case encrypt.SensitiveClassification:
						// TODO
					case encrypt.UnknownClassification:
						// TODO
					case "":
						pass.Reportf(fieldRef.Pos(), "empty classification string")
					default:
						pass.Reportf(fieldRef.Pos(), "invalid data classification: %q", strValue)
					}
				default:
					// TODO: implement other types?
				}
			case *ssa.Call: // dynamic string
				// TODO: resolve dynamic strings
			default:
				// TODO: implement other types?
			}
		}
	case "Filter":
		for _, fieldRef := range *fieldRefs {
			fieldRefStore, ok := (fieldRef).(*ssa.Store)
			if !ok {
				continue
			}
			switch value := fieldRefStore.Val.(type) {
			case *ssa.Const: // static string
				switch value.Value.Kind() {
				case constant.String:
					strValue, err := strconv.Unquote(value.Value.ExactString())
					if err != nil {
						return fmt.Errorf("failed to unquote classification field value string: %w", err)
					}
					switch encrypt.FilterOperation(strValue) {
					case encrypt.RedactOperation:
						// TODO
					case encrypt.EncryptOperation:
						// TODO
					case encrypt.HmacSha256Operation:
						// TODO
					case encrypt.NoOperation:
						// TODO
					default:
						pass.Reportf(fieldRef.Pos(), "invalid filter operation: %q", strValue)
					}
				default:
					// TODO: implement other types?
				}
			case *ssa.Call: // dynamic string
				// TODO: resolve dynamic strings
			default:
				// TODO: implement other types?
			}
		}
	default:
		// Assumes no other fields exist
	}

	return nil
}

func checkCompositLiteral(pass *analysis.Pass, builtSSA *buildssa.SSA, cl *ast.CompositeLit) error {
	for _, e := range cl.Elts {
		// TODO: check for required key->value pairs
		switch et := e.(type) {
		case *ast.KeyValueExpr:
			switch fmt.Sprintf("%v", et.Key) {
			case "Pointer":
				v := pass.TypesInfo.Types[et.Value]
				switch fmt.Sprintf("%v", v.Value) {
				case `""`:
					pass.Reportf(et.Value.Pos(), "empty pointerstructure pointer string")
				default:
					// TODO
				}
			case "Classification":
				v := pass.TypesInfo.Types[et.Value]
				operation := fmt.Sprintf("%v", v.Value)
				operation = strings.TrimPrefix(operation, `"`)
				operation = strings.TrimSuffix(operation, `"`)
				switch encrypt.DataClassification(operation) {
				case encrypt.PublicClassification:
					// TODO
				case encrypt.SecretClassification:
					// TODO
				case encrypt.SensitiveClassification:
					// TODO
				case encrypt.UnknownClassification:
					// TODO
				default:
					pass.Reportf(et.Value.Pos(), "invalid data classification: %q", operation)
				}
			case "Filter":
				v := pass.TypesInfo.Types[et.Value]
				operation := fmt.Sprintf("%v", v.Value)
				operation = strings.TrimPrefix(operation, `"`)
				operation = strings.TrimSuffix(operation, `"`)
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
					pass.Reportf(et.Value.Pos(), "invalid filter operation: %q", operation)
				}
			}
		default:
			// not implemented
		}
	}
	return nil
}
