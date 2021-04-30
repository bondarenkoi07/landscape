package Model

//Cтруктура операции
type ChangeData struct {
	Mode  string `json:"mode"`
	X     int64  `json:"x"`
	Y     int64  `json:"y"`
	Model string `json:"model"`
}
