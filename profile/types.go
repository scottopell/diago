package profile

import "fmt"

type StringsMap map[uint64]string
type LocationsMap map[uint64]Location
type FunctionsMap map[uint64]Function
type ManyFunctionsMap map[uint64][]Function // TODO(remy): better naming...

type Sample struct {
	Functions    []Function
	Value        int64
	PercentTotal float64
}

type Samples []Sample

type Location struct {
	Functions []Function
}

type Function struct {
	Name       string `json:"name"`
	File       string `json:"file"`
	LineNumber uint64 `json:"lineNumber"`
	Self       int64  `json:"self"`
}

type SampleMode string

var (
	// use this when you don't really know the mode
	// to use to read the profile.
	ModeDefault   SampleMode = ""
	ModeCpu       SampleMode = "cpu"
	ModeHeapAlloc SampleMode = "heap-alloc"
	ModeHeapInuse SampleMode = "heap-inuse"
)

func (f Function) String(lineNumber bool) string {
	if lineNumber {
		return fmt.Sprintf("%s %s:%d", f.Name, f.File, f.LineNumber)
	}
	return fmt.Sprintf("%s %s", f.Name, f.File)
}
