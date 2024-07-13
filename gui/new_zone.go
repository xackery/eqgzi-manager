package gui

// NewZone is a struct populated with data obtained from the dialog
type NewZone struct {
	Name string
}

// ShowNewZone opens a form to create a new zone
func ShowNewZone() (*NewZone, error) {
	return &NewZone{}, nil
}
