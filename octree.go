package octree

import (
	"errors"
	"sync"
)

var (
	// ErrNotFound is returned by Remove if an entity was not found in the tree
	ErrNotFound = errors.New("Entity not found")
)

// The Octree type defines an octree of entities that can be efficiently locked seperatly and altered concurrently
type Octree struct {
	root   *octant
	levels uint8

	leafMapLock sync.Locker
	leafMap     map[interface{}]*leaf
}

type node interface {
	find(x, y, z, sqDist float64) <-chan interface{}
}

type octant struct {
	// Optional lock
	lock sync.Locker

	// The children nodes of this node.
	// For which octant is which, see: http://en.wikipedia.org/wiki/Octant_(solid_geometry)
	children [8]node

	// The X, Y, Z coordinates of the center of this octant
	x, y, z float64
	// The size of this octant
	size float64

	// The level of the octant, mostly kept for ease of debugging
	level uint8
}

type leaf struct {
	lock     sync.Locker
	children map[interface{}][3]float64

	// The X, Y, Z coordinates of the center of this octant
	x, y, z float64
	// The size of this octant
	size float64
}

// New creates a new Octree using the passed configuration
// preGenLevels levels in the tree will be preGenerated so they can be used without locking of any kind
func New(x, y, z, size float64, totalLevels, preGenLevels uint8) (*Octree, error) {
	if preGenLevels > totalLevels {
		return nil, errors.New("preGenLevels may not be bigger than totalLevels")
	}

	if size <= 0 {
		return nil, errors.New("Size must be a positive number greater than zero")
	}

	t := &Octree{
		leafMap:     make(map[interface{}]*leaf),
		leafMapLock: &sync.Mutex{},
		levels:      totalLevels,
		root: &octant{
			x:    x,
			y:    y,
			z:    z,
			size: size,
		}}

	var preGen func(oct *octant, currLevel uint8)
	preGen = func(oct *octant, currLevel uint8) {
		for i := 0; i < 8; i++ {
			if currLevel != totalLevels {
				oct.children[i] = oct.newOctant(i, currLevel == preGenLevels)
				if currLevel < preGenLevels {
					preGen(oct.children[i].(*octant), currLevel+1)
				}
			} else {
				oct.children[i] = oct.newLeaf(i)
			}
		}
	}

	preGen(t.root, 0)

	return t, nil
}

// getChildInfo computes x, y, z coordinates and a size for the child of the octant in the given quadrant
func (oct *octant) getChildInfo(i int) (x, y, z, size float64) {
	// Calculate coordinates
	if i&1 != 1 {
		x = oct.x + oct.size/2
	} else {
		x = oct.x - oct.size/2
	}

	if i&2 != 2 {
		y = oct.y + oct.size/2
	} else {
		y = oct.y - oct.size/2
	}

	if i&4 != 4 {
		z = oct.z + oct.size/2
	} else {
		z = oct.z - oct.size/2
	}

	size = oct.size / 2
	return
}

func (oct *octant) newOctant(i int, locked bool) *octant {
	child := &octant{}
	if locked {
		child.lock = &sync.Mutex{}
	}

	child.x, child.y, child.z, child.size = oct.getChildInfo(i)
	child.level = oct.level + 1

	return child
}

func (oct *octant) newLeaf(i int) *leaf {
	leaf := &leaf{
		lock:     &sync.Mutex{},
		children: make(map[interface{}][3]float64),
	}

	leaf.x, leaf.y, leaf.z, leaf.size = oct.getChildInfo(i)

	return leaf
}
