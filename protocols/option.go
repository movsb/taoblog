package protocols

type Option struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type GetOptionRequest struct {
	Name    string
	Default bool
	Value   string
}

type ListOptionsRequest struct {
}

type ListOptionsResponse struct {
	Options []*Option `json:"options"`
}
