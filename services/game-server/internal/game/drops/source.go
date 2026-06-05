package drops

type SourceType string

const SourceTypeAsteroid SourceType = "asteroid"

type Source struct {
	Type SourceType
	ID   string
	Size int
	X    float64
	Y    float64
}
