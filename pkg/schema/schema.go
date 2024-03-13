package schema

import (
	"net/url"
	"reflect"
)

type Schema struct {
	typeDefs TypeDefs
}

type TypeDefs map[any]*Definition

type Definition struct {
	handler          IType
	currentNameIndex int
	names            ParameterNames
}

type ParameterNames []string

func NewSchema(typeDefs TypeDefs) *Schema {
	defs := TypeDefs{}

	for typ, value := range typeDefs {
		var handler IType

		switch typ {
		case Integer:
			handler = IntegerType{}
		case Float:
			handler = FloatType{}
		case String:
			handler = StringType{}
		case ArrayInteger:
			handler = ArrayIntegerType{}
		case ArrayString:
			handler = ArrayStringType{}
		case Boolean:
			handler = BooleanType{}
		case Duration:
			handler = DurationType{}
		case Struct:
			handler = StructType{}
		}

		defs[typ] = &Definition{
			names: ParameterNames{},
		}

		defs[typ].handler = handler

		if value != nil {
			defs[typ].names = value.names
		}
	}

	return &Schema{typeDefs: defs}
}

func NewType(names ParameterNames) *Definition {
	def := &Definition{
		names: names,
	}

	return def
}

func (s *Schema) ConvertToQuery(query url.Values, properties ...any) {
	for _, property := range properties {
		typeOf, definition, name := reflect.TypeOf(property), new(Definition), ""

		switch typeOf.Kind() {
		case reflect.Struct:
			definition = s.typeDefs[Struct]
		default:
			definition = s.typeDefs[typeOf.String()]
		}

		if definition != nil {
			if currentIndex := definition.currentNameIndex; len(definition.names) > definition.currentNameIndex {
				indexName := definition.names[currentIndex]

				if !query.Has(indexName) {
					name = indexName
				} else if len(definition.names) > currentIndex+1 {
					name = definition.names[currentIndex+1]
				}

				definition.currentNameIndex++
			}

			definition.handler.doQuery(query, name, property)
		}
	}
}
