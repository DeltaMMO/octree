package octree

import "testing"

func TestSetZero(t *testing.T) {
	o, err := New(0, 0, 0, 100, 3, 0)
	if err != nil {
		t.Errorf("Error while creating tree: %v", err)
		t.FailNow()
	}

	// Simplistic 0,0,0 test
	d := &struct{}{}
	if err := o.Set(d, 0, 0, 0); err != nil {
		t.Errorf("Error while setting tree value: %v", err)
		t.FailNow()
	}

	curr := o.root.children[0].(*octant)
	for i := 0; i < 2; i++ {
		curr = curr.children[7].(*octant)
	}

	if curr.children[7].(*leaf).children[d] != [3]float64{0, 0, 0} {
		t.Errorf("Error while checking tree consistency: Unexpected entity position %v", curr.children[7].(*leaf).children[d])
	}
}

func TestRemove(t *testing.T) {
	o, err := New(0, 0, 0, 100, 3, 0)
	if err != nil {
		t.Errorf("Error while creating tree: %v", err)
		t.FailNow()
	}

	d := &struct{}{}
	if o.Set(d, 0, 0, 0); err != nil {
		t.Errorf("Error while setting tree value: %v", err)
		t.FailNow()
	}

	if err := o.Remove(d); err != nil {
		t.Errorf("Error while removing from tree: %v", err)
		t.FailNow()
	}

	curr := o.root.children[0].(*octant)
	for i := 0; i < 2; i++ {
		curr = curr.children[7].(*octant)
	}

	if _, ok := curr.children[7].(*leaf).children[d]; ok {
		t.Error("Error while checking tree consistency: Deleted entry was present")
	}
}

func TestRemoveNotFound(t *testing.T) {
	o, err := New(0, 0, 0, 100, 3, 0)
	if err != nil {
		t.Errorf("Error while creating tree: %v", err)
		t.FailNow()
	}

	if err := o.Remove(&struct{}{}); err != ErrNotFound {
		t.Errorf("Unexpected result when removing non-existent item: %v", err)
		t.FailNow()
	}

}
