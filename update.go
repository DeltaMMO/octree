package octree

// Set will add a new entity or update an existing one at the specified position
func (t *Octree) Set(entity interface{}, x, y, z float64) error {
	// Figure out where to put the entity
	// We do this before locking the whole thing so writing can still be reasonably parallel
	leaf := t.getOrCreateLeaf(x, y, z)

	// (Re)add the entity, we do this while having the leafMapLock locked
	// this way we can prevent a potential deadlock situation where two Set calls
	// have eachother's leaf locked
	t.leafMapLock.Lock()
	defer t.leafMapLock.Unlock()

	if err := t.remove(entity); err != nil && err != ErrNotFound {
		return err
	}

	leaf.lock.Lock()
	defer leaf.lock.Unlock()
	leaf.children[entity] = [3]float64{x, y, z}
	t.leafMap[entity] = leaf

	return nil
}

// Remove will remove the passed entity from the tree, it returns ErrNotFound if the entity is not present
func (t *Octree) Remove(entity interface{}) error {
	t.leafMapLock.Lock()
	defer t.leafMapLock.Unlock()

	return t.remove(entity)
}

// remove is the actual private implementation of the Remove function, it assumes t.leafMapLock has been taken
func (t *Octree) remove(entity interface{}) error {
	if leaf, valid := t.leafMap[entity]; valid {
		leaf.lock.Lock()
		defer leaf.lock.Unlock()

		delete(leaf.children, entity)
		delete(t.leafMap, entity)
		return nil
	}

	return ErrNotFound
}

// Get the leaf where a certain point is located
// creating octants and the leaf itself on the way there if necessary
func (t *Octree) getOrCreateLeaf(x, y, z float64) *leaf {
	return t.root.getOrCreateLeaf(x, y, z, 0, t.levels)
}

func (oct *octant) getOrCreateLeaf(x, y, z float64, level, maxLevels uint8) *leaf {
	if oct.lock != nil {
		oct.lock.Lock()
	}

	// Figure out which side we're on
	// X
	var i int
	if x < oct.x {
		i |= 1
	}

	// Y
	if y < oct.y {
		i |= 2
	}

	// Z
	if z < oct.z {
		i |= 4
	}

	node := oct.children[i]
	if level == maxLevels {
		// We're handling leaves now
		if node == nil {
			node = oct.newLeaf(i)
			oct.children[i] = node
		}

		// Unlock before bailing out
		if oct.lock != nil {
			oct.lock.Unlock()
		}
		return node.(*leaf)
	}

	// Handle more octants
	if node == nil {
		node = oct.newOctant(i, true)
		oct.children[i] = node
	}

	// We're done reading/modifying this particular octant, we can release our lock before recursing
	if oct.lock != nil {
		oct.lock.Unlock()
	}
	return node.(*octant).getOrCreateLeaf(x, y, z, level+1, maxLevels)

}
