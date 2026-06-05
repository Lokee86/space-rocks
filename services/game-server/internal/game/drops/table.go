package drops

type DropMode string

const (
	DropModeSingle DropMode = "single"
	DropModeMulti  DropMode = "multi"
)

type Entry struct {
	PickupType    string
	Chance        float64
	MinSourceSize int
	MaxSourceSize int
}

type Table struct {
	ID               string
	SourceType       SourceType
	DropMode         DropMode
	MaxDropsPerSource int
	MaxActivePickups int
	Entries          []Entry
}

type Tables struct {
	ByID map[string]Table
}

type Roll struct {
	Values []float64
}

type Result struct {
	TableID    string
	PickupType string
	X          float64
	Y          float64
}

func (tables Tables) Evaluate(tableID string, source Source, roll Roll) ([]Result, bool) {
	results := tables.Roll(tableID, source, roll)
	return results, len(results) > 0
}
