package service

import "encoding/json"

const (
	defaultPageSize = 100
)

type Paginator struct {
	Offset int `json:"offset"`
	Size   int `json:"size"`
}

func newPaginator(offset, size int) *Paginator {
	return &Paginator{offset, size}
}

func parsePage(data string) (*Paginator, error) {
	var page Paginator
	err := json.Unmarshal([]byte(data), &page)
	if err != nil {
		return nil, err
	}
	return newPaginator(page.Offset, page.Size), nil
}

func (p *Paginator) MustEncode() string {
	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(b)
}
