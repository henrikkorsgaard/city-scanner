package experiment

type Experiment struct {
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	Zoom      int     `json:"zoom"`
	Database  string
}

type Node struct {
	Id        string
	Latitude  float64
	Longitude float64
}
