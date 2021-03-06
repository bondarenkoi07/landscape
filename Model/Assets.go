package Model

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

type Models struct {
	models []string
}

func NewModels(files []os.FileInfo) Models {
	var newModels Models
	for _, file := range files {
		if !file.IsDir() {
			err := newModels.Append(file.Name())
			if err != nil {
				log.Println(err)
			}
		}
	}
	return newModels
}

func (A *Models) Append(value string) error {
	matched, err := regexp.MatchString(`^(.*)\.png$`, value)
	if err == nil {
		if matched {
			(*A).models = append((*A).models, value)
			return nil
		} else {
			return errors.New("ModelAddException: " + value + " file might be not a png")
		}
	} else {
		return err
	}
}

func (A *Models) Find(value string) bool {
	for _, item := range (*A).models {
		if value == item {
			return true
		}
	}
	return false
}

func CheckModel(msg ChangeData) bool {
	files, err := ioutil.ReadDir("static/assets/models")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	validator := NewModels(files)

	if !validator.Find(msg.Model) && (msg.Mode == "DeleteModel" || msg.Mode == "PlaceModel") {
		return false
	}
	return true
}
