package main

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"regexp"
	"strings"

	// Generated files - mandatory for (custom) options.
	option "github.com/geomodular/meta-store/gen/ai/h2o/meta_store"
)

const (
	contextPackage      = protogen.GoImportPath("context")
	arangoDriverPackage = protogen.GoImportPath("github.com/arangodb/go-driver")
	assetPackage        = protogen.GoImportPath("github.com/geomodular/meta-store/pkg/server/asset")
	artifactPackage     = protogen.GoImportPath("github.com/geomodular/meta-store/pkg/artifact")
	errorsPackage       = protogen.GoImportPath("github.com/pkg/errors")
	grpcPackage         = protogen.GoImportPath("google.golang.org/grpc")
	protoPackage        = protogen.GoImportPath("github.com/geomodular/meta-store/pkg/proto")
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			genAll(gen, f)
		}
		return nil
	})
}

func genAll(plugin *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {

	filename := file.GeneratedFilenamePrefix + ".pb.meta.go"
	g := plugin.NewGeneratedFile(filename, file.GoImportPath)

	genHeader(g, file)

	for _, service := range file.Services {
		genService(g, service)
	}

	for _, message := range file.Messages {
		genMessage(g, message)
	}

	return g
}

func genHeader(g *protogen.GeneratedFile, file *protogen.File) {
	g.P("// Code generated by protoc-gen-meta. DO NOT EDIT!!!")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
}

func genService(g *protogen.GeneratedFile, service *protogen.Service) {

	// NOTE: To load extension ahead check: https://github.com/golang/protobuf/issues/1260
	options := service.Desc.Options().(*descriptorpb.ServiceOptions)

	serviceName := strings.TrimSuffix(service.GoName, "Service")
	bigServiceName := capitalize(serviceName)
	smallServiceName := lower(serviceName)
	serviceStructName := fmt.Sprintf("%sServiceServer", smallServiceName)
	collectionName := proto.GetExtension(options, option.E_CollectionName).(string)

	g.P("func init() {")
	g.P(assetPackage.Ident("RegisterGRPCInitializer"), "(initGRPC)")
	g.P(assetPackage.Ident("RegisterCollection"), "(\"", collectionName, "\")")
	g.P("}")

	g.P()

	g.P("func initGRPC(grpcServer *", grpcPackage.Ident("Server"), ", db ", arangoDriverPackage.Ident("Database"), ") {")
	g.P(smallServiceName, "Server := New", bigServiceName, "ServiceServer(db)")
	g.P("Register", bigServiceName, "ServiceServer(grpcServer, ", smallServiceName, "Server)")
	g.P("}")

	g.P()

	g.P("type ", serviceStructName, " struct {")
	g.P("db ", arangoDriverPackage.Ident("Database"))
	g.P("}")

	g.P()

	g.P("func New", bigServiceName, "ServiceServer(db ", arangoDriverPackage.Ident("Database"), ") *", serviceStructName, " {")
	g.P("return &", serviceStructName, "{db: db}")
	g.P("}")

	for _, method := range service.Methods {
		mName := method.GoName
		mType := determineMethodType(mName)
		mInput := method.Input.GoIdent
		mOutput := method.Output.GoIdent

		// TODO: bigServiceName != Meta equivalent
		// TODO: what to do with plural?

		g.P("func (x *", serviceStructName, ") ", mName, "(ctx ", contextPackage.Ident("Context"), ", req *", mInput, ") (*", mOutput, ", error) {")
		switch mType {
		case CREATE:
			g.P("inArtifact := NewMeta", bigServiceName, "FromProto(req.Get", bigServiceName, "())")
			g.P("outArtifact, err := ", artifactPackage.Ident("Create"), "[Meta", bigServiceName, "](ctx, x.db, \"", collectionName, "\", inArtifact)")
			g.P("if err != nil { return nil, ", errorsPackage.Ident("Wrap"), "(err, \"failed creating ", bigServiceName, "\") }")
			g.P("return outArtifact.ToProto(), nil")
		case GET:
			g.P("resourceName := req.GetName()")
			g.P("outArtifact, err := ", artifactPackage.Ident("Get"), "[Meta", bigServiceName, "](ctx, x.db, \"", collectionName, "\", resourceName)")
			g.P("if err != nil { return nil, ", errorsPackage.Ident("Wrap"), "(err, \"failed getting ", bigServiceName, "\") }")
			g.P("return outArtifact.ToProto(), nil")
		case LIST:
			g.P("inToken := req.GetPageToken()")
			g.P("inSize := int(req.GetPageSize())")
			g.P("outToken, outSize, outArtifacts, err := ", artifactPackage.Ident("List"), "[Meta", bigServiceName, "](ctx, x.db, \"", collectionName, "\", inToken, inSize)")
			g.P("if err != nil { return nil, ", errorsPackage.Ident("Wrap"), "(err, \"failed listing ", bigServiceName, "\") }")
			g.P("var artifacts []*", bigServiceName)
			g.P("for _, a := range outArtifacts { artifacts = append(artifacts, a.ToProto()) }")
			g.P("return &", mOutput, "{ NextPageToken: outToken, TotalSize: int32(outSize), ", bigServiceName, "s: artifacts }, nil")
		case REMOVE:
			g.P("resourceName := req.GetName()")
			g.P("err := artifact.Delete(ctx, x.db, \"", collectionName, "\", resourceName)")
			g.P("if err != nil { return nil, ", errorsPackage.Ident("Wrap"), "(err, \"failed removing", bigServiceName, "\") }")
			g.P("return &", mOutput, "{}, nil")
		case UPDATE:
			g.P("inArtifact := NewMeta", bigServiceName, "FromProto(req.Get", bigServiceName, "())")
			g.P("outArtifact, err := artifact.Update[Meta", bigServiceName, "](ctx, x.db, \"", collectionName, "\", inArtifact.Name, inArtifact)")
			g.P("if err != nil { return nil, ", errorsPackage.Ident("Wrap"), "(err, \"failed updating ", bigServiceName, "\") }")
			g.P("return outArtifact.ToProto(), nil")
		default:
			g.P("panic(\"not implemented\")")
		}
		g.P("}")
	}
}

func genMessage(g *protogen.GeneratedFile, message *protogen.Message) {

	options := message.Desc.Options().(*descriptorpb.MessageOptions)
	collectionType := proto.GetExtension(options, option.E_CollectionType).(option.CollectionType)

	if collectionType == option.CollectionType_UNDEFINED {
		return // Messages that don't have collection type are not processed.
	}

	genMetaStruct(g, message)
	genAccessors(g, message)
	genProtoConversions(g, message)
}

func genMetaStruct(g *protogen.GeneratedFile, message *protogen.Message) {
	g.P("type Meta", message.GoIdent, " struct {")
	g.P()
	g.P("// Mandatory `key` field.")
	g.P("Key string `json:\"_key,omitempty\"`")
	g.P()
	g.P("// Other fields.")
	for _, field := range message.Fields {
		t, p := fieldGoType(g, field)
		if p {
			g.P(field.GoName, " *", t, " `json:\"", toArangoIdent(field.GoName), "\"`")
		} else {
			g.P(field.GoName, " ", t, " `json:\"", toArangoIdent(field.GoName), "\"`")
		}
	}
	g.P("}")
	g.P()
}

func genAccessors(g *protogen.GeneratedFile, message *protogen.Message) {
	g.P("func (x *Meta", message.GoIdent, ") SetKey(value string) {")
	g.P("x.Key = value")
	g.P("}")
	for _, field := range message.Fields {
		t, _ := fieldGoType(g, field)
		switch t {
		case "string":
			g.P("func (x *Meta", message.GoIdent, ") Set", field.GoName, "(value string) {")
			g.P("x.", field.GoName, " = value")
			g.P("}")
		default:
		}
	}
	g.P()
}

func genProtoConversions(g *protogen.GeneratedFile, message *protogen.Message) {
	g.P("func (x *Meta", message.GoIdent, ") ToProto() *", message.GoIdent, "{")
	g.P("return &", message.GoIdent, "{")
	for _, field := range message.Fields {
		t, _ := fieldGoType(g, field)
		switch t {
		case "time.Time":
			g.P(field.GoName, ": ", protoPackage.Ident("ToProtoTimestamp"), "(x.", field.GoName, "),")
		default:
			g.P(field.GoName, ": x.", field.GoName, ",")
		}
	}
	g.P("}")
	g.P("}")
	g.P()
	g.P("func NewMeta", message.GoIdent, "FromProto(dataset *", message.GoIdent, ") *Meta", message.GoIdent, "{")
	g.P("return &Meta", message.GoIdent, "{")
	g.P("Key: \"\",")
	for _, field := range message.Fields {
		t, _ := fieldGoType(g, field)
		switch t {
		case "time.Time":
			g.P(field.GoName, ": ", protoPackage.Ident("FromProtoTimestamp"), "(dataset.", field.GoName, "),")
		default:
			g.P(field.GoName, ": dataset.", field.GoName, ",")
		}
	}
	g.P("}")
	g.P("}")
	g.P()
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// https://stackoverflow.com/questions/56616196/how-to-convert-camel-case-string-to-snake-case
func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func capitalize(str string) string {
	return strings.ToUpper(str[0:1]) + str[1:]
}

func lower(str string) string {
	return strings.ToLower(str[0:1]) + str[1:]
}

func toArangoIdent(name string) string {
	snake := toSnakeCase(name)
	return strings.ToLower(snake)
}

func fieldGoType(g *protogen.GeneratedFile, field *protogen.Field) (goType string, pointer bool) {
	if field.Desc.IsWeak() {
		return "struct{}", false
	}

	pointer = field.Desc.HasPresence()
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		goType = "bool"
	case protoreflect.EnumKind:
		goType = g.QualifiedGoIdent(field.Enum.GoIdent)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		goType = "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		goType = "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		goType = "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		goType = "uint64"
	case protoreflect.FloatKind:
		goType = "float32"
	case protoreflect.DoubleKind:
		goType = "float64"
	case protoreflect.StringKind:
		goType = "string"
	case protoreflect.BytesKind:
		goType = "[]byte"
		pointer = false // rely on nullability of slices for presence
	case protoreflect.MessageKind:
		if field.Message.GoIdent.GoName == "Timestamp" {
			g.QualifiedGoIdent(protogen.GoIdent{"", "time"})
			goType = "time.Time"
		} else {
			goType = "*" + g.QualifiedGoIdent(field.Message.GoIdent)
		}
		pointer = false
	case protoreflect.GroupKind:
		goType = "*" + g.QualifiedGoIdent(field.Message.GoIdent)
		pointer = false // pointer captured as part of the type
	}

	switch {
	case field.Desc.IsList():
		return "[]" + goType, false
	case field.Desc.IsMap():
		keyType, _ := fieldGoType(g, field.Message.Fields[0])
		valType, _ := fieldGoType(g, field.Message.Fields[1])
		return fmt.Sprintf("map[%v]%v", keyType, valType), false
	}

	return goType, pointer
}

type methodType int

const (
	UNKNOWN methodType = iota
	CREATE
	GET
	LIST
	REMOVE
	UPDATE
)

func determineMethodType(name string) methodType {
	if strings.HasPrefix(name, "Create") {
		return CREATE
	} else if strings.HasPrefix(name, "Get") {
		return GET
	} else if strings.HasPrefix(name, "List") {
		return LIST
	} else if strings.HasPrefix(name, "Remove") {
		return REMOVE
	} else if strings.HasPrefix(name, "Update") {
		return UPDATE
	}
	return UNKNOWN
}
