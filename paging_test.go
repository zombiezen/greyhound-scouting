package main

import (
	"os"
	"reflect"
	"testing"
)

type intSlicePager struct {
	slice  []int
	offset int
	limit  int
}

func newIntSlicePager(data ...int) *intSlicePager {
	return &intSlicePager{
		slice: data,
		limit: len(data),
	}
}

func (pager *intSlicePager) Count() (int, os.Error) {
	return len(pager.slice), nil
}

func (pager *intSlicePager) Offset(n int) Pager {
	pager.offset = n
	return pager
}

func (pager *intSlicePager) Limit(n int) Pager {
	pager.limit = n
	return pager
}

func (pager *intSlicePager) All(in interface{}) os.Error {
	ptr := in.(*[]int)
	switch {
	case pager.offset > len(pager.slice):
		*ptr = nil
	case pager.offset+pager.limit > len(pager.slice):
		*ptr = pager.slice[pager.offset:]
	default:
		*ptr = pager.slice[pager.offset : pager.offset+pager.limit]
	}
	return nil
}

func TestPageZero(t *testing.T) {
	p, err := NewPaginator(newIntSlicePager(), 10)
	if err != nil {
		t.Fatalf("NewPaginator error: %v", err)
	}

	if p.NPage() != 1 {
		t.Errorf("p.NPage() != 1 (got %v)", p.NPage())
	}
	if p.Page(0) != nil {
		t.Error("Got non-nil for p.Page(0)")
	}
	if p.Page(2) != nil {
		t.Error("Got non-nil for p.Page(2)")
	}

	if p1 := p.Page(1); p1 != nil {
		if p1.HasNext() {
			t.Error("p.Page(1).HasNext() is wrong")
		}
		if p1.HasPrevious() {
			t.Error("p.Page(1).HasPrevious() is wrong")
		}
		if n := p1.Number(); n != 1 {
			t.Errorf("p.Page(1).Number() != 1 (got %v)", n)
		}

		var result []int
		if err := p1.Get(&result); err != nil {
			t.Errorf("p.Page(1).Get(...) error: %v", err)
		}
		if len(result) != 0 {
			t.Error("Page results are non-nil")
		}
	} else {
		t.Error("Got nil for p.Page(1)")
	}
}

func TestNPage(t *testing.T) {
	if p, err := NewPaginator(newIntSlicePager(42, -7, 98, 100, 5, 4), 2); err == nil {
		if n := p.NPage(); n != 3 {
			t.Errorf("NPage of 6 items returns %d (expected 3)", n)
		}
	} else {
		t.Errorf("NewPaginator error: %v", err)
	}

	if p, err := NewPaginator(newIntSlicePager(42, -7, 98, 100, 5), 2); err == nil {
		if n := p.NPage(); n != 3 {
			t.Errorf("NPage of 5 items returns %d (expected 3)", n)
		}
	} else {
		t.Errorf("NewPaginator error: %v", err)
	}
}

func TestPageHasNext(t *testing.T) {
	p, err := NewPaginator(newIntSlicePager(42, -7, 98, 100, 5), 2)
	if err != nil {
		t.Fatalf("NewPaginator error: %v", err)
	}
	if n := p.NPage(); n != 3 {
		t.Fatalf("NPage of 5 items returns %d (expected 3)", n)
	}

	if p1 := p.Page(1); p1 != nil {
		if !p1.HasNext() {
			t.Error("Page(1).HasNext() is wrong")
		}
	} else {
		t.Error("Page(1) is nil")
	}

	if p2 := p.Page(2); p2 != nil {
		if !p2.HasNext() {
			t.Error("Page(2).HasNext() is wrong")
		}
	} else {
		t.Error("Page(2) is nil")
	}

	if p3 := p.Page(3); p3 != nil {
		if p3.HasNext() {
			t.Error("Page(3).HasNext() is wrong")
		}
	} else {
		t.Error("Page(3) is nil")
	}
}

func TestPageHasPrevious(t *testing.T) {
	p, err := NewPaginator(newIntSlicePager(42, -7, 98, 100, 5), 2)
	if err != nil {
		t.Fatalf("NewPaginator error: %v", err)
	}
	if n := p.NPage(); n != 3 {
		t.Fatalf("NPage of 5 items returns %d (expected 3)", n)
	}

	if p1 := p.Page(1); p1 != nil {
		if p1.HasPrevious() {
			t.Error("Page(1).HasPrevious() is wrong")
		}
	} else {
		t.Error("Page(1) is nil")
	}

	if p2 := p.Page(2); p2 != nil {
		if !p2.HasPrevious() {
			t.Error("Page(2).HasPrevious() is wrong")
		}
	} else {
		t.Error("Page(2) is nil")
	}

	if p3 := p.Page(3); p3 != nil {
		if !p3.HasPrevious() {
			t.Error("Page(3).HasPrevious() is wrong")
		}
	} else {
		t.Error("Page(3) is nil")
	}
}

func TestPageGet(t *testing.T) {
	p, err := NewPaginator(newIntSlicePager(42, -7, 98, 100, 5), 2)
	if err != nil {
		t.Fatalf("NewPaginator error: %v", err)
	}
	if n := p.NPage(); n != 3 {
		t.Fatalf("NPage of 5 items returns %d (expected 3)", n)
	}

	checkPageGet(t, p, 1, []int{42, -7})
	checkPageGet(t, p, 2, []int{98, 100})
	checkPageGet(t, p, 3, []int{5})
}

func checkPageGet(t *testing.T, p *Paginator, num int, expected []int) {
	page := p.Page(num)
	if page == nil {
		t.Errorf("p.Page(%d) is nil", num)
		return
	}
	var vals []int
	if err := page.Get(&vals); err != nil {
		t.Errorf("p.Page(%d).Get(...) error: %v", num, err)
		return
	}
	if !reflect.DeepEqual(vals, expected) {
		t.Errorf("p.Page(%d).Get(...) != %#v (got %#v)", num, expected, vals)
	}
}
