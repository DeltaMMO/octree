package octree

import (
	"math"
	"testing"
)

func TestNewSanityChecks(t *testing.T) {
	o, err := New(0, 0, 0, 1, 1, 2)
	if err == nil || o != nil {
		t.Error("Expected error when creating an octree with too many preGenLevels")
	}

	o, err = New(0, 0, 0, 0, 0, 0)
	if err == nil || o != nil {
		t.Error("Expected error when creating an octree with size 0")
	}

	o, err = New(0, 0, 0, -5, 0, 0)
	if err == nil || o != nil {
		t.Error("Expected error when creating an octree with negative size")
	}
}

func TestPregen(t *testing.T) {
	o, err := New(0, 0, 0, 100, 2, 2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if o.root == nil {
		t.Error("Root is not defined")
		t.FailNow()
	}

	var check func(oct *octant, level uint8)
	check = func(oct *octant, level uint8) {
		for i := 0; i < 8; i++ {
			expectedSize := 100 / (math.Pow(2, float64(level)))
			if oct.size != expectedSize {
				t.Errorf("Consistency check failed on level %d: Unexpected size %.2f, expected %.2f", level, oct.size, expectedSize)
				t.FailNow()
			}

			if level < 2 {
				if oct.lock != nil {
					t.Errorf("Consistency check failed on level %d: Unexpected lock on pregen'd level", level)
					t.FailNow()
				}

				if oct.children[i] == nil {
					t.Errorf("Constency check failed on level %d: Pregen'd level is nil", level)
					t.FailNow()
				}
				check(oct.children[i].(*octant), level+1)
			} else {
				if _, ok := oct.children[i].(*leaf); !ok {
					t.Errorf("Consistency check failed on level %d: final level is not a leaf", level)
					t.FailNow()
				}
			}
		}
	}

	check(o.root, 0)
}
