package pagination

import (
	"errors"
	"fmt"
	"math"
	"net/url"
)

// PageMeta contains the pagination meta data.
type PageMeta struct {
	Limit  int
	Offset int
}

// PageLinks contains methods for pagination links.
type PageLinks interface {
	FirstPageMeta() PageMeta
	FirstPageLink() (string, error)
	HasPrevPage() bool
	PrevPageMeta() PageMeta
	PrevPageLink() (string, error)
	HasNextPage() bool
	NextPageMeta() PageMeta
	NextPageLink() (string, error)
	LastPageMeta() PageMeta
	LastPageLink() (string, error)
	ToHeader() (string, error)
}

// NewPageLinks creates a new pageLinks object.
func NewPageLinks(baseUrl, rawQuery string, total, limit, offset int) (PageLinks, error) {
	l := pageLinks{
		baseUrl:  baseUrl,
		pageMeta: BuildPageMeta(total, limit, offset),
		rawQuery: rawQuery,
	}
	return l, nil
}

// pageLinks is a PageLinks implementation.
type pageLinks struct {
	baseUrl  string
	pageMeta map[string]PageMeta
	rawQuery string
}

// ToHeader returns a header field representation of the links.
func (p pageLinks) ToHeader() (string, error) {
	var link string
	var err error
	header := ""

	// first page
	link, err = p.FirstPageLink()
	if err != nil {
		return "", err
	}
	header += fmt.Sprintf(`<%s>; rel="first"`, link)

	// prev page
	if p.HasPrevPage() {
		link, err = p.PrevPageLink()
		if err != nil {
			return "", err
		}
		header += fmt.Sprintf(`, <%s>; rel="prev"`, link)
	}

	// next page
	if p.HasNextPage() {
		link, err = p.NextPageLink()
		if err != nil {
			return "", err
		}
		header += fmt.Sprintf(`, <%s>; rel="next"`, link)
	}

	// last page
	link, err = p.LastPageLink()
	if err != nil {
		return "", err
	}
	header += fmt.Sprintf(`, <%s>; rel="last"`, link)

	return header, nil
}

// FirstPageMeta returns the metadata of the first page.
func (p pageLinks) FirstPageMeta() PageMeta {
	return p.pageMeta["first"]
}

// FirstPageLink returns the link to the first page.
func (p pageLinks) FirstPageLink() (string, error) {
	return buildLink(p.baseUrl, p.rawQuery, p.FirstPageMeta().Offset)
}

// PrevPageMeta returns the metadata of the previous page.
func (p pageLinks) PrevPageMeta() PageMeta {
	return p.pageMeta["prev"]
}

// HasPrevPage returns true when there is a previous page.
func (p pageLinks) HasPrevPage() bool {
	_, ok := p.pageMeta["prev"]
	return ok
}

// PrevPageLink returns the link to the previous page.
func (p pageLinks) PrevPageLink() (string, error) {
	if !p.HasPrevPage() {
		return "", fmt.Errorf("Pagination has no previous page")
	}
	return buildLink(p.baseUrl, p.rawQuery, p.PrevPageMeta().Offset)
}

// NextPageMeta returns the metadata of the next page.
func (p pageLinks) NextPageMeta() PageMeta {
	return p.pageMeta["next"]
}

// HasNextPage returns true when there is a next page.
func (p pageLinks) HasNextPage() bool {
	_, ok := p.pageMeta["next"]
	return ok
}

// NextPageLink returns the link to the next page.
func (p pageLinks) NextPageLink() (string, error) {
	if !p.HasNextPage() {
		return "", fmt.Errorf("Pagination has no next page")
	}
	return buildLink(p.baseUrl, p.rawQuery, p.NextPageMeta().Offset)
}

// LastPageMeta returns the metadata of the last page.
func (p pageLinks) LastPageMeta() PageMeta {
	return p.pageMeta["last"]
}

// LastPageLink returns the link to the last page.
func (p pageLinks) LastPageLink() (string, error) {
	return buildLink(p.baseUrl, p.rawQuery, p.LastPageMeta().Offset)
}

// BuildPageMeta build the metadata for the provided pagination coordinates.
func BuildPageMeta(total int, limit int, offset int) map[string]PageMeta {
	meta := map[string]PageMeta{
		"first": {Limit: limit, Offset: 0},
		"last":  {Limit: limit, Offset: getLastOffset(total, limit)},
	}
	if offset > 0 && offset < total && offset-limit >= 0 {
		meta["prev"] = PageMeta{Limit: limit, Offset: max(0, offset-limit)}
	}
	if total >= (limit + offset) {
		meta["next"] = PageMeta{Limit: limit, Offset: offset + limit}
	}
	return meta
}

// max returns the maximum of the provided values.
func max(x int, y int) int {
	if x > y {
		return x
	}
	return y
}

// getLastOffset returns the last offset for the total/limit values.
func getLastOffset(total, limit int) int {
	offset := (total/limit - 1) * limit
	if total%limit != 0 {
		offset += limit
	}
	return offset
}

// buildLink build the link for the provided values.
func buildLink(baseUrl string, rawQuery string, offset int) (string, error) {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", err
	}
	values.Set("offset", fmt.Sprintf("%d", offset))
	if baseUrl[len(baseUrl)-1:] != "?" {
		baseUrl += "?"
	}
	return baseUrl + values.Encode(), nil
}

const (
	minLimit = 1
	maxLimit = 1000
)

// SortParserFunc is a function to parse a sort string.
type SortParserFunc func(sort string) (string, error)

// Pagination contains pagination information.
type Pagination struct {
	Limit  int
	Offset int
	Sort   string
}

// Validate validates to Pagination object.
func (p *Pagination) Validate() error {
	if p.Limit < 1 {
		return errors.New("limit must be positive")
	}
	if p.Offset < 0 {
		return errors.New("offset cannot be negative")
	}
	return nil
}

func (p *Pagination) String() string {
	return fmt.Sprintf("Limit: %d, Offset: %d, Sort: %s", p.Limit, p.Offset, p.Sort)
}

// NewPaginationFromArgs returns a new Pagination object based on the provided pagination arguments.
func NewPaginationFromArgs(limit, offset *int, sort *string) (Pagination, error) {
	p := Pagination{}

	if limit == nil {
		p.Limit = 50
	} else {
		if *limit < minLimit || *limit > maxLimit {
			return p, fmt.Errorf("limit must have a value between %d and %d", minLimit, maxLimit)
		}
		p.Limit = *limit
	}

	if offset != nil && *offset > 0 {
		p.Offset = *offset
	}

	if sort != nil {
		p.Sort = *sort
	}

	return p, nil
}

func DefaultPagination() Pagination {
	return Builder().Build()
}

func Builder() *builder {
	b := &builder{
		p: Pagination{Limit: 100, Offset: 0},
	}
	return b
}

type builder struct {
	p Pagination
}

func (b *builder) WithLimit(limit int) *builder {
	b.p.Limit = limit
	return b
}

func (b *builder) WithOffset(offset int) *builder {
	b.p.Offset = offset
	return b
}

func (b *builder) WithSort(sort string) *builder {
	b.p.Sort = sort
	return b
}

func (b *builder) Build() Pagination {
	return b.p
}

func All() Pagination {
	return Builder().WithLimit(math.MaxInt).WithOffset(0).Build()
}

func One() Pagination {
	return Builder().WithLimit(1).WithOffset(0).Build()
}

// PageCount calculates to total number of pages.
func PageCount(count int, limit int) int {
	if count == 0 {
		return 1
	}
	if count%limit != 0 {
		return count/limit + 1
	}
	return count / limit
}
