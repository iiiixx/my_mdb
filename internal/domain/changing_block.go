package domain

type ChangingBlock struct {
	Kind   string      `json:"kind"`
	Title  string      `json:"title"`
	Movies []MovieCard `json:"movies"`
}
