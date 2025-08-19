package packageA

type TestSource struct {
	Name          string
	Age           int
	Other         string `gen-copier:"Another"`
	SourceExtra   string
	DifferentType int
}
