package model

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	gojs "github.com/xeipuuv/gojsonschema"

	category "github.com/veganbase/backend/services/category-service/client"
	"github.com/veganbase/backend/services/item-service/model/schema"
)

//go:generate go-bindata -pkg schema -o schema/schema_data_generated.go -ignore schema_data_generated.go -prefix schema/ schema/...

// Schema is a representation of a JSON schema, as a map from field
// names to values.
type Schema map[string]interface{}

// SchemaMap is a map from schema names to schemas.
type SchemaMap map[string]*gojs.Schema

// Schema map.
var schemas SchemaMap

// Validate validates JSON data against a named schema.
func Validate(schema string, data []byte) (*gojs.Result, error) {
	if schemas == nil {
		return nil, errors.New("schema map has not been initialised")
	}

	s, ok := schemas[schema]
	if !ok {
		return nil, errors.New("unknown JSON schema '" + schema + "'")
	}

	return s.Validate(gojs.NewBytesLoader(data))
}

// LoadSchemas reads and expands all JSON validation schemas, resolves
// references to utility schemas and compiles the resulting schemas
// for validation. Must be called before Validate is used. Not
// thread-safe! Call it from main before server is started.
func LoadSchemas(cat category.Client) {
	// Are they already loaded?
	if schemas != nil {
		return
	}

	// Add category format checkers.
	for name, checker := range LoadCategoryCheckers(cat) {
		gojs.FormatCheckers.Add(name, checker)
	}

	// Read and expand all schemas.
	inSchemas, utilSchemas := readRawSchemas()
	expandSchemas(inSchemas)

	// Load and compile all schemas, adding utility schemas to loader.
	ss := SchemaMap{}
	for name, s := range inSchemas {
		loader := gojs.NewSchemaLoader()
		loader.AddSchemas(utilSchemas...)
		compiled, err := loader.Compile(gojs.NewGoLoader(s))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to compile JSON schema '" + name + "'")
		}
		ss[name] = compiled
	}

	schemas = ss
}

// Read raw JSON schemas from compiled-in bindata storage. These come
// in two varieties: the item schemas and "utility" schemas referenced
// in the item schemas. These two cases are treated differently, so
// they're returned separately here.
func readRawSchemas() (map[string]Schema, []gojs.JSONLoader) {
	schemas := map[string]Schema{}
	utilSchemas := []gojs.JSONLoader{}
	for _, n := range schema.AssetNames() {
		raw, err := schema.Asset(n)
		if err != nil {
			log.Fatal().Err(err).
				Str("schema", n).
				Msg("couldn't read compiled-in JSON schema")
		}

		schema := Schema{}
		err = json.Unmarshal(raw, &schema)
		if err != nil {
			log.Fatal().Err(err).
				Str("schema", n).
				Msg("couldn't unmarshal compiled-in JSON schema")
		}

		if strings.HasPrefix(n, "utils/") {
			utilSchemas = append(utilSchemas, gojs.NewGoLoader(schema))
			continue
		}

		title, ok := schema["title"]
		if !ok {
			log.Fatal().Err(err).
				Str("schema", n).
				Msg("couldn't find title for compiled-in JSON schema")
		}
		stitle, ok := title.(string)
		if !ok {
			log.Fatal().Err(err).
				Str("schema", n).
				Msg("wrong type for JSON schema title")
		}
		schemas[stitle] = schema
	}
	return schemas, utilSchemas
}

// Expand "extends" relationships between JSON schemas.
func expandSchemas(schemas map[string]Schema) {
	// Make set of names of all schemas with "extends" field.
	remaining := map[string]bool{}
	for n, s := range schemas {
		_, extends := s["extends"]
		if extends {
			remaining[n] = true
		}
	}

	// While there are still schemas needing expansion...
	for len(remaining) > 0 {
		progress := false

		// Try to expand each remaining schema that needs expansion.
		for n := range remaining {
			// Look up the schema and get the schema it extends from.
			toExtend := schemas[n]
			es, chk := toExtend["extends"].(string)
			if !chk {
				log.Fatal().Msg("invalid type for 'extends' in '" + n + "' JSON schema")
			}
			extendWith, ok := schemas[es]
			if !ok {
				log.Fatal().Msg("unknown schema '" + es + "' in extends for '" +
					n + "' JSON schema")
			}

			// The schema we're going to extend from must already have been
			// extended.
			_, chk = extendWith["extends"]
			if chk {
				continue
			}

			// We have a schema we want to extend and a schema to extend it
			// from, so we can make progress.
			progress = true
			extendSchema(toExtend, extendWith)
			delete(remaining, n)
		}

		// If we didn't make any progress on this round, that's an error.
		if !progress {
			r := []string{}
			for n := range remaining {
				r = append(r, n)
			}
			log.Fatal().
				Str("remaining", strings.Join(r, ",")).
				Msg("failed to make progress in JSON schema expansion")
		}
	}
}

func extendSchema(toExtend Schema, extendWith Schema) {
	toName := toExtend["title"].(string)
	withName := extendWith["title"].(string)

	// Run through the properties in the extendWith schema, check that
	// they aren't already in the toExtend schema, then add them to the
	// properties of the toExtend schema.
	fromPropsOrig, ok := extendWith["properties"]
	if !ok {
		log.Fatal().Msg("schema '" + withName + "' being extended from doesn't have properties")
	}
	fromProps, ok := fromPropsOrig.(map[string]interface{})
	if !ok {
		log.Fatal().Msg("invalid properties field in schema '" + withName + "' being extended from")
	}
	toPropsOrig, ok := toExtend["properties"]
	if !ok {
		log.Fatal().Msg("schema '" + toName + "' being extended doesn't have properties")
	}
	toProps, ok := toPropsOrig.(map[string]interface{})
	if !ok {
		log.Fatal().Msg("invalid properties field in schema '" + toName + "' being extended")
	}
	for name, prop := range fromProps {
		_, chk := toProps[name]
		if chk {
			log.Fatal().Msg("duplicate property '" + name + "' in schema '" + toName + "' extension")
		}
		toProps[name] = prop
	}

	// Add the required properties from the extendWith schema to the
	// toExtend schema.
	fromReqOrig, ok := extendWith["required"]
	if !ok {
		log.Fatal().Msg("schema being extended from doesn't have required")
	}
	fromReq, ok := fromReqOrig.([]interface{})
	if !ok {
		log.Fatal().Msg("invalid required field in schema being extended from")
	}
	toReqOrig, ok := toExtend["required"]
	if !ok {
		log.Fatal().Msg("schema being extended doesn't have required")
	}
	toReq, ok := toReqOrig.([]interface{})
	if !ok {
		log.Fatal().Msg("invalid required field in schema being extended")
	}
	toExtend["required"] = append(toReq, fromReq...)

	// Remove the extends field from the toExtend schema.
	delete(toExtend, "extends")
}
