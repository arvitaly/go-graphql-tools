package test

//Package for declare struct in some packages

type C struct {
	CExt
}
type CExt struct {
	Str2   string  `json:"str2"`
	Int1   int     `json:"int1"`
	Float1 float64 `json:"float1"`
}
