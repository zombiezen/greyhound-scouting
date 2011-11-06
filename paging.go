package main

import (
	"os"
	"launchpad.net/mgo"
)

// A Pager is something that can be paginated over.
type Pager interface {
	Count() (int, os.Error)
	Offset(int) Pager
	Limit(int) Pager
	All(interface{}) os.Error
}

type Paginator struct {
	Pager   Pager
	PerPage int
	count   int
}

func NewPaginator(pager Pager, per int) (*Paginator, os.Error) {
	if per < 1 {
		return nil, os.NewError("There must be at least one result per page")
	}
	count, err := pager.Count()
	if err != nil {
		return nil, err
	}
	return &Paginator{
		Pager:   pager,
		PerPage: per,
		count:   count,
	}, nil
}

// NPage returns the number of pages.
func (paginator *Paginator) NPage() int {
	if paginator.count == 0 {
		return 1
	}
	return (paginator.count + paginator.PerPage - 1) / paginator.PerPage
}

// Page returns the 1-based page.
func (paginator *Paginator) Page(n int) *Page {
	if n < 1 || n > paginator.NPage() {
		return nil
	}
	return &Page{paginator, n - 1}
}

type Page struct {
	*Paginator
	Index int
}

// Number returns the 1-based page number.
func (page Page) Number() int {
	return page.Index + 1
}

func (page Page) HasNext() bool {
	return page.Number() < page.NPage()
}

func (page Page) HasPrevious() bool {
	return page.Number() > 1
}

// Get fetches all of the objects on the page.
func (page Page) Get(i interface{}) os.Error {
	return page.Pager.Offset(page.Index * page.PerPage).Limit(page.PerPage).All(i)
}

// MongoPager wraps an mgo query so that it can be used as a Pager.
type MongoPager struct {
	*mgo.Query
}

func (pager MongoPager) Limit(n int) Pager {
	return MongoPager{pager.Query.Limit(n)}
}

func (pager MongoPager) Offset(n int) Pager {
	return MongoPager{pager.Query.Skip(n)}
}
