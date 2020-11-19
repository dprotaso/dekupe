package main

import (
	"io"
	"os"

	. "github.com/dprotaso/go-yit" //nolint: stylecheck // Allow this dot import.
	"gopkg.in/yaml.v3"
)

type obj struct {
	apiVersion string
	kind       string
	name       string
	namespace  string
}

func (o *obj) IsValid() bool {
	// we allow empty namespaces (cluster scope)
	return o.apiVersion != "" && o.kind != "" && o.name != ""

}

func main() {
	decoder := yaml.NewDecoder(os.Stdin)

	e := yaml.NewEncoder(os.Stdout)
	e.SetIndent(2)
	defer e.Close()

	set := make(map[obj]struct{})

	for {
		var doc yaml.Node
		if err := decoder.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		obj := parse(&doc)

		if obj.IsValid() {
			if _, ok := set[obj]; ok {
				continue
			}
			set[obj] = struct{}{}
		}

		if err := e.Encode(&doc); err != nil {
			panic(err)
		}
	}
}

func parse(doc *yaml.Node) obj {
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		doc = doc.Content[0]
	}

	if doc.Tag == "!!null" {
		return obj{}
	}

	return obj{
		apiVersion: apiVersion(doc),
		kind:       kind(doc),
		name:       name(doc),
		namespace:  namespace(doc),
	}
}

func valueOrEmpty(it Iterator) string {
	node, ok := it()
	if !ok {
		return ""
	}

	return node.Value
}

func apiVersion(doc *yaml.Node) string {
	it := FromNode(doc).
		ValuesForMap(
			// key predicate
			WithStringValue("apiVersion"),
			// value predicate
			StringValue,
		)

	return valueOrEmpty(it)
}

func kind(doc *yaml.Node) string {
	it := FromNode(doc).
		ValuesForMap(
			// key predicate
			WithStringValue("kind"),
			//value predicate
			StringValue,
		)

	return valueOrEmpty(it)
}

func name(doc *yaml.Node) string {
	it := FromNode(doc).
		ValuesForMap(
			// key predicate
			WithStringValue("metadata"),
			//value predicate
			WithKind(yaml.MappingNode),
		).
		ValuesForMap(
			// key predicate
			WithStringValue("name"),
			//value predicate
			StringValue,
		)

	return valueOrEmpty(it)
}

func namespace(doc *yaml.Node) string {
	it := FromNode(doc).
		ValuesForMap(
			// key predicate
			WithStringValue("metadata"),
			//value predicate
			WithKind(yaml.MappingNode),
		).
		ValuesForMap(
			// key predicate
			WithStringValue("namespace"),
			//value predicate
			StringValue,
		)

	return valueOrEmpty(it)
}
