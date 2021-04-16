package Model

type Transaction struct {
	Landscape map[int][]int    `json:"landscape"`
	Object    map[int][]string `json:"object"`
	Texture   string           `json:"texture"`
}
