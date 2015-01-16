package octree

import "sync"

// FindChan looks for all the entities within dist of the passed point
// The output is a channel that will be fed each entity as it is found
func (t *Octree) FindChan(findX, findY, findZ, dist float64) <-chan interface{} {
	return t.root.find(findX, findY, findZ, dist*dist)
}

// FindSlice looks for all the entities within dist of the passed point
// The output is a slice that will contain all entities that were found
func (t *Octree) FindSlice(findX, findY, findZ, dist float64) []interface{} {
	var output []interface{}

	for entity := range t.FindChan(findX, findY, findZ, dist) {
		output = append(output, entity)
	}

	return output
}

// GetPosition looks up the position of the passed entity in the tree
// This requires a global lock
func (t *Octree) GetPosition(entity interface{}) (x, y, z float64) {
	t.leafMapLock.Lock()
	defer t.leafMapLock.Unlock()
	leaf := t.leafMap[entity]
	x, y, z = leaf.x, leaf.y, leaf.z
	return
}

func (oct *octant) find(findX, findY, findZ, sqDist float64) <-chan interface{} {
	output := make(chan interface{})

	go func() {
		defer close(output)

		var subChans []<-chan interface{}

		if oct.lock != nil {
			oct.lock.Lock()
		}

		// Go through each axis in the quadrant
		// This is easily done by looking from 0 to 7 and the 3 numbers at binary level
		// each number in binary is mapped to positive or negative for that axis
		for i := 0; i < 8; i++ {
			var maxX, maxY, maxZ, minX, minY, minZ float64

			// Calculate each extremum of the bounding cube for this quadrant
			// X
			if i&1 != 1 {
				maxX = oct.x + oct.size
				minX = oct.x
			} else {
				maxX = oct.x
				minX = oct.x - oct.size
			}

			// Y
			if i&2 != 2 {
				maxY = oct.y + oct.size
				minY = oct.y
			} else {
				maxY = oct.y
				minY = oct.y - oct.size
			}

			// Z
			if i&4 != 4 {
				maxZ = oct.z + oct.size
				minY = oct.z
			} else {
				maxZ = oct.z
				minZ = oct.z - oct.size
			}

			// Collision check between a sphere and a cube, see http://stackoverflow.com/questions/4578967/cube-sphere-intersection-test
			sqD := sqDist
			var v float64

			// X
			if findX < minX {
				v = findX - minX
			} else if findX > maxX {
				v = findX - maxX
			}
			sqD -= v * v
			v = 0

			// Y
			if findY < minY {
				v = findY - minY
			} else if findY > maxY {
				v = findY - maxY
			}
			sqD -= v * v
			v = 0

			// Z
			if findZ < minZ {
				v = findZ - minZ
			} else if findZ > maxZ {
				v = findZ - maxZ
			}
			sqD -= v * v
			v = 0

			if sqD > 0 {
				// Start the searches through the lower levels in parallel
				child := oct.children[i]
				if child != nil {
					subChans = append(subChans, child.find(findX, findY, findZ, sqDist))
				}
			}
		}

		// We're done with this node, now we just need to wait for the subnodes
		// First things first: Unlock it so other stuff can happen on it
		if oct.lock != nil {
			oct.lock.Unlock()
		}

		if len(subChans) > 0 {
			// Get the result from each of the subChans and pipe it into our own output channel
			var wg sync.WaitGroup
			wg.Add(len(subChans))

			// Start an output goroutine for each input channel. out copies values from c to out until c is closed, then calls wg.Done.
			out := func(c <-chan interface{}) {
				for n := range c {
					output <- n
				}
				wg.Done()
			}

			for _, c := range subChans {
				go out(c)
			}

			wg.Wait()
		}
	}()

	return output
}

func (l *leaf) find(findX, findY, findZ, sqDist float64) <-chan interface{} {
	output := make(chan interface{})

	go func() {
		l.lock.Lock()
		defer l.lock.Unlock()
		defer close(output)

		// Loop through each entity and do a distance check
		for data, pos := range l.children {
			xd := findX - pos[0]
			yd := findY - pos[1]
			zd := findZ - pos[2]

			if xd*xd+yd*yd+zd*zd < sqDist {
				output <- data
			}
		}
	}()

	return output
}
