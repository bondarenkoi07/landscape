package Model

import (
	"errors"
	"fmt"
	"io/ioutil"
)

const (
	MaxHeight    = 3
	MinHeight    = 0
	CanvasHeight = 600
	CanvasWidth  = 1200
	GridCount    = 30
)

type Transaction struct {
	Landscape map[int64][]int64  `json:"landscape"`
	Object    map[int64][]string `json:"object"`
	Texture   string             `json:"texture"`
}

//TODO: create Transaction initializer

func (t Transaction) Check(msg ChangeData) (bool, error) {
	row, rowExists := t.Landscape[msg.Y]
	if !rowExists {
		return false, errors.New(fmt.Sprintf("wrong y %d", msg.Y))
	}

	length := int64(len(row))

	if length < msg.X || msg.X < 0 {
		return false, errors.New(fmt.Sprintf("wrong x %d, length: %d", msg.Y, length))
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

func (t Transaction) CheckMap() error {
	files, err := ioutil.ReadDir("static/assets/models")
	if err != nil {
		return err
	}

	validator := NewModels(files)
	for i := range t.Object {
		for j := range t.Object[i] {
			item := t.Object[i][j]

			if item != "" {

				if !validator.Find(item) {
					return err
				}
			}
		}
	}

	for i := range t.Landscape {
		for j := range t.Landscape[i] {
			item := t.Landscape[i][j]
			if item > MaxHeight || item < MinHeight {
				return errors.New("invalid height")
			}
		}
	}
	return nil
}
