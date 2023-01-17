package meta

type Type string

// Type has same name as string for neater comparisons.
const (
	Bool     Type = "Bool"
	Duration Type = "Duration"
	Float    Type = "Float"
	Integer  Type = "Integer"
	List     Type = "List"
	Text     Type = "Text"
)

// Data is the specification for a Metadata
type Data struct {
	Aliases       []string
	Tags          []string
	Label         string
	Description   string
	Documentation string
	Default       string
	Options       []string
}

// Metadata is a set of accessor functions that never write to the store and
// thus do not create race conditions.
type Metadata struct {
	Aliases       func() []string
	Tags          func() []string
	Label         func() string
	Description   func() string
	Documentation func() string
	Default       func() string
	Options       func() []string
	Typ           Type
}

// New loads Data into a Metadata.
func New(d Data, t Type) Metadata {
	return Metadata{
		func() []string { return d.Aliases },
		func() []string { return d.Tags },
		func() string { return d.Label },
		func() string { return d.Description },
		func() string { return d.Documentation },
		func() string { return d.Default },
		func() []string { return d.Options },
		t,
	}
}
