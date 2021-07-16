package store

import "testing"

func TestStore(t *testing.T)  {
	s := NewStore()
	s.Add("foo", "bar")
	if item, ok := s.Get("foo"); !ok{
		t.Errorf("didn't find inserted item")
	}else {
		if item != "bar"{
			t.Errorf("except \" bar\", but is: %s", item)
		}
	}

	s.Update("foo", "baz")

	if item, ok := s.Get("foo"); !ok{
		t.Errorf("didn't find inserted item")
	}else {
		if item != "baz"{
			t.Errorf("except \" baz\", but is: %s", item)
		}
	}

	// test delete

	s.Delete("foo")

	if _, ok := s.Get("foo"); ok{
		t.Errorf("should find inserted item??")
	}
	// test list

	s.Add("a", "a1")
	s.Add("b", "b1")
	s.Add("c", "c1")

	a := s.List()
	//arr1 := []string{"a1", "b1","c1"}
	if len(a) != 3{
		t.Errorf("not three?")
	}

}