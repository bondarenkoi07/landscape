package Model

import "errors"

const (
	MaxHeight = 3
	MinHeight = 0
)

type Transaction struct {
	Landscape map[int64][]int64  `json:"landscape"`
	Object    map[int64][]string `json:"object"`
	Texture   string             `json:"texture"`
}

func (t Transaction) Check(msg ChangeData) (bool, error) {
	row, rowExists := t.Landscape[msg.Y]
	if !rowExists {
		return false, errors.New("wrong y")
	}

	length := int64(len(row))

	if length > msg.X {
		return false, errors.New("wrong x")
	}
	height := row[msg.X]

	switch msg.Mode {
	case "up":
		if height+1 > MaxHeight {
			return false, errors.New("max height reached")
		}
	case "down":
		if height-1 < MinHeight {
			return false, errors.New("min height reached")
		}

	case "PlaceModel":
		if t.Object[msg.Y][msg.X] != "" {
			return false, errors.New("model has already placed")
		}

		if !CheckModel(msg) {
			return false, errors.New("unknown model")
		}
	default:
		return false, nil

	}
	return true, nil
}
