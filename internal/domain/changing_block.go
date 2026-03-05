package domain

type ChangingBlock struct {
	Kind   string
	Title  string
	Movies []Movie
}

type ChangingBlockResponse struct {
	Kind   string      `json:"kind"`
	Title  string      `json:"title"`
	Movies []MovieCard `json:"movies"`
}
