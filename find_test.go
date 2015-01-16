package octree

import "testing"

func TestFind(t *testing.T) {
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

	data := o.FindSlice(0, 0, 0, 1)
	if len(data) != 1 || data[0] != d {
		t.Error("Did not find added element in tree")
		t.FailNow()
	}
}

func TestGetPosition(t *testing.T) {
	o, err := New(0, 0, 0, 100, 3, 0)
	if err != nil {
		t.Errorf("Error while creating tree: %v", err)
		t.FailNow()
	}

	d := &struct{}{}
	if err := o.Set(d, 5, 20, 90); err != nil {
		t.Errorf("Error while setting tree value: %v", err)
		t.FailNow()
	}

	x, y, z := o.GetPosition(d)

}
