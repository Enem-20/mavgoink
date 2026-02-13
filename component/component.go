package component

type Component struct {
	ID   byte   `json:"id"`
	Name string `json:"name"`
}

func NewComponent(id byte, name string) *Component {
	return &Component{ID: id, Name: name}
}
