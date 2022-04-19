package client

type PageQuery struct {
	Current int `json:"current"`
	Size    int `json:"size"`
}

func NewPageQuery(current, size int) *PageQuery {
	return &PageQuery{
		Current: current,
		Size:    size,
	}
}

func (p *PageQuery) Source() (r map[string]interface{}, err error) {
	if p.Size <= 0 || p.Current <= 0 {
		return
	}
	r = make(map[string]interface{})
	r["current"] = p.Current
	r["size"] = p.Size
	return
}
