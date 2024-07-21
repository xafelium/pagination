package pagination

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestBuildPageMeta(t *testing.T) {
	tests := []struct {
		total    int
		limit    int
		offset   int
		expected map[string]PageMeta
	}{
		{
			total: 5, limit: 20, offset: 20,
			expected: map[string]PageMeta{
				"first": {20, 0},
				"last":  {20, 0},
			},
		},
		{
			total: 40, limit: 20, offset: 20,
			expected: map[string]PageMeta{
				"first": {20, 0},
				"prev":  {20, 0},
				"next":  {20, 40},
				"last":  {20, 20},
			},
		},
		{
			total: 79, limit: 20, offset: 60,
			expected: map[string]PageMeta{
				"first": {20, 0},
				"prev":  {20, 40},
				"last":  {20, 60},
			},
		},
		{
			total: 100, limit: 20, offset: 60,
			expected: map[string]PageMeta{
				"first": {20, 0},
				"prev":  {20, 40},
				"next":  {20, 80},
				"last":  {20, 80},
			},
		},
		{
			total: 100, limit: 20, offset: 0,
			expected: map[string]PageMeta{
				"first": {20, 0},
				"next":  {20, 20},
				"last":  {20, 80},
			},
		},
		{
			total: 100, limit: 20, offset: 60,
			expected: map[string]PageMeta{
				"first": {20, 0},
				"prev":  {20, 40},
				"next":  {20, 80},
				"last":  {20, 80},
			},
		},
		{
			total: 100, limit: 16, offset: 32,
			expected: map[string]PageMeta{
				"first": {16, 0},
				"prev":  {16, 16},
				"next":  {16, 48},
				"last":  {16, 96},
			},
		},
		{
			total: 102, limit: 20, offset: 60,
			expected: map[string]PageMeta{
				"first": {20, 0},
				"prev":  {20, 40},
				"next":  {20, 80},
				"last":  {20, 100},
			},
		},
		{
			total: 200, limit: 20, offset: 0,
			expected: map[string]PageMeta{
				"first": {20, 0},
				"next":  {20, 20},
				"last":  {20, 180},
			},
		},
		{
			total: 200, limit: 20, offset: 20,
			expected: map[string]PageMeta{
				"first": {20, 0},
				"prev":  {20, 0},
				"next":  {20, 40},
				"last":  {20, 180},
			},
		},
	}

	for _, tc := range tests {
		meta := BuildPageMeta(tc.total, tc.limit, tc.offset)
		if !reflect.DeepEqual(tc.expected, meta) {
			t.Errorf("BuildPageMeta(%d, %d, %d):\n\texpected: %+v\n\tgot: %+v",
				tc.total, tc.limit, tc.offset, tc.expected, meta)
		}
	}
}

func TestGetLastOffset(t *testing.T) {
	tests := []struct {
		total    int
		limit    int
		expected int
	}{
		{total: 100, limit: 5, expected: 95},
		{total: 200, limit: 10, expected: 190},
		{total: 150, limit: 7, expected: 147},
		{total: 199, limit: 9, expected: 198},
		{total: 100, limit: 15, expected: 90},
	}

	for _, tc := range tests {
		offset := getLastOffset(tc.total, tc.limit)
		if offset != tc.expected {
			t.Errorf("getLastOffset(total: %d, limit: %d):\n\texpected: %+v\n\tgot: %+v",
				tc.total, tc.limit, tc.expected, offset)
		}
	}
}

func TestNew(t *testing.T) {
	baseUrl := "https://api.example.com/api/v1/cars"
	rawQuery := "offset=60&limit=15"
	total := 100
	limit := 15
	offset := 60

	p, err := NewPageLinks(baseUrl, rawQuery, total, limit, offset)
	if err != nil {
		t.Error(err)
	}

	firstPageLink, err := p.FirstPageLink()
	if err != nil {
		t.Fatalf("FirstPageLink() error: %+v", err)
	}
	expectedFirstPageLink := "https://api.example.com/api/v1/cars?limit=15&offset=0"
	if firstPageLink != expectedFirstPageLink {
		t.Fatalf("FirstPageLink()\n\texpected: %s\n\tgot: %s",
			expectedFirstPageLink, firstPageLink)
	}

	prevPageLink, err := p.PrevPageLink()
	if err != nil {
		t.Fatalf("PrevPageLink() error: %+v", err)
	}
	expectedPrevPageLink := "https://api.example.com/api/v1/cars?limit=15&offset=45"
	if prevPageLink != expectedPrevPageLink {
		t.Fatalf("PrevPageLink()\n\texpected: %s\n\tgot: %s",
			expectedPrevPageLink, prevPageLink)
	}

	nextPageLink, err := p.NextPageLink()
	if err != nil {
		t.Fatalf("NextPageLink() error: %+v", err)
	}
	expectedNextPageLink := "https://api.example.com/api/v1/cars?limit=15&offset=75"
	if nextPageLink != expectedNextPageLink {
		t.Fatalf("NextPageLink()\n\texpected: %s\n\tgot: %s",
			expectedNextPageLink, nextPageLink)
	}

	lastPageLink, err := p.LastPageLink()
	if err != nil {
		t.Fatalf("LastPageLink() error: %+v", err)
	}
	expectedLastPageLink := "https://api.example.com/api/v1/cars?limit=15&offset=90"
	if lastPageLink != expectedLastPageLink {
		t.Fatalf("LastPageLink()\n\texpected: %s\n\tgot: %s",
			expectedLastPageLink, lastPageLink)
	}
}

func TestToHeader(t *testing.T) {
	expectedHeader :=
		"<http://www.example.com/abc?limit=10&offset=0>; rel=\"first\", " +
			"<http://www.example.com/abc?limit=10&offset=40>; rel=\"prev\", " +
			"<http://www.example.com/abc?limit=10&offset=60>; rel=\"next\", " +
			"<http://www.example.com/abc?limit=10&offset=100>; rel=\"last\""
	baseUrl := "http://www.example.com/abc"
	rawQuery := "offset=50&limit=10"
	total := 110
	limit := 10
	offset := 50

	p, err := NewPageLinks(baseUrl, rawQuery, total, limit, offset)
	if err != nil {
		t.Error(err)
	}

	actualHeader, err := p.ToHeader()
	if err != nil {
		t.Fatal(err)
	}
	if actualHeader != expectedHeader {
		t.Fatalf("ToHeader():\n\texpected: %s\n\t     got: %s",
			expectedHeader, actualHeader)
	}
}

func TestPageCount(t *testing.T) {
	tests := []struct {
		count             int
		limit             int
		expectedPageCount int
	}{
		{count: 0, limit: 1, expectedPageCount: 1},
		{count: 0, limit: 2, expectedPageCount: 1},
		{count: 0, limit: 3, expectedPageCount: 1},
		{count: 1, limit: 1, expectedPageCount: 1},
		{count: 1, limit: 2, expectedPageCount: 1},
		{count: 1, limit: 3, expectedPageCount: 1},
		{count: 2, limit: 1, expectedPageCount: 2},
		{count: 2, limit: 2, expectedPageCount: 1},
		{count: 2, limit: 3, expectedPageCount: 1},
		{count: 3, limit: 1, expectedPageCount: 3},
		{count: 3, limit: 2, expectedPageCount: 2},
		{count: 3, limit: 3, expectedPageCount: 1},
		{count: 3, limit: 4, expectedPageCount: 1},
		{count: 4, limit: 1, expectedPageCount: 4},
		{count: 4, limit: 2, expectedPageCount: 2},
		{count: 4, limit: 3, expectedPageCount: 2},
		{count: 4, limit: 4, expectedPageCount: 1},
		{count: 4, limit: 5, expectedPageCount: 1},
		{count: 5, limit: 1, expectedPageCount: 5},
		{count: 5, limit: 2, expectedPageCount: 3},
		{count: 5, limit: 3, expectedPageCount: 2},
		{count: 5, limit: 4, expectedPageCount: 2},
		{count: 5, limit: 5, expectedPageCount: 1},
		{count: 5, limit: 6, expectedPageCount: 1},
		{count: 6, limit: 1, expectedPageCount: 6},
		{count: 6, limit: 2, expectedPageCount: 3},
		{count: 6, limit: 3, expectedPageCount: 2},
		{count: 6, limit: 4, expectedPageCount: 2},
		{count: 6, limit: 5, expectedPageCount: 2},
		{count: 6, limit: 6, expectedPageCount: 1},
		{count: 6, limit: 7, expectedPageCount: 1},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("count: %d, limit: %d", test.count, test.limit), func(t *testing.T) {
			require.Equal(t, test.expectedPageCount, PageCount(test.count, test.limit))
		})
	}
}
