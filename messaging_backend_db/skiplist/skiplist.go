// This package implements the skiplist interface and defines each of its methods.
package skiplist

import (
	"cmp"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"sync/atomic"
)

// UpdateCheck is a function used to determine what a specific object should do when it is trying to insert an existing node into a skiplist
type UpdateCheck[K cmp.Ordered, V any] func(key K, currValue V, exists bool) (newValue V, err error)

// Pair is a key, value pair that is used in the return of the Query function
type Pair[K cmp.Ordered, V any] struct {
	Key   K
	Value V
}

// SkipList is an interface with an Upsert, Remove, Find, and Query function
type SkipList[K cmp.Ordered, V any] interface {
	Upsert(key K, check UpdateCheck[K, V]) (updated bool, err error)
	Remove(key K) (removedValue V, removed bool)
	Find(key K) (foundValue V, found bool)
	Query(start K, end K) (results []Pair[K, V], err error)
}

// A node in the concurrent list. Stores a key and a corresponding item.
type node[K cmp.Ordered, V any] struct {
	mtx sync.Mutex

	key      K
	value    V
	topLevel int

	// whether the node has been removed from the list. defaults to false.
	marked atomic.Bool

	fullyLinked atomic.Bool

	// A list of pointers to next nodes, where the index in the list corresponds to the level
	next []atomic.Pointer[node[K, V]]
}

// A list data structure that supports concurrent access. The list's nodes
// contain key-value pairs. The list may contain at most one node with a given
// key. Also keeps track of the maxlevel that any node can have, and an atomic timestamp counter for Query
type List[K cmp.Ordered, V any] struct {
	head      *node[K, V]
	tail      *node[K, V]
	maxlevel  int
	timestamp atomic.Int32 // could be 64, arbitrary choice
}

// Constructs and returns a new list that supports concurrent access.
// `minKey` must be strictly smaller and `maxKey` must be strictly greater than
// any keys that will be stored in the list. If this assumption is violated,
// the behaviour of the list is unspecified.
func NewList[K cmp.Ordered, V any](minKey K, maxKey K) List[K, V] {

	// using maxlevel = 5 for placeholder:
	maxlev := 5

	slog.Info("in new list")

	var dummyHead node[K, V] = node[K, V]{
		key:  minKey,
		next: make([]atomic.Pointer[node[K, V]], maxlev),
	}
	var dummyTail node[K, V] = node[K, V]{
		key:  maxKey,
		next: make([]atomic.Pointer[node[K, V]], maxlev),
	}

	slog.Info("before assiging head and tail")

	for i := 0; i < maxlev; i++ {
		dummyHead.next[i].Store(&dummyTail)
		dummyTail.next[i].Store(&dummyHead)
	}

	slog.Info("before return")

	return List[K, V]{
		head:      &dummyHead,
		tail:      &dummyTail,
		maxlevel:  maxlev - 1,
		timestamp: atomic.Int32{},
	}
}

// Helper function for Find. Takes in a key and returns the level at which the key was found and
// a list of preceding and succeeding nodes. On failure, this returns -1, nil, nil
func (s *List[K, V]) findHelp(key K) (int, []*node[K, V], []*node[K, V]) {

	slog.Info("starting findhelp")
	fmt.Println(key)
	// before anything, make sure this key is larger than head and smaller than tail
	// if key<=head.key, then that means key is empty. can't have empty stuff
	if key <= s.head.key {
		return -1, nil, nil
	} else if key >= s.tail.key {
		// hopefully this just always makes the tail's key larger than anything in else
		s.tail.key = key + key
	}

	foundLevel := -1
	level := s.maxlevel
	// pred := s.head[level].Load()
	pred := s.head
	var curr *node[K, V]

	// make a list of nodes of length level
	preds := make([]*node[K, V], level+1)
	succs := make([]*node[K, V], level+1)

	slog.Info("before for loop of find help")

	for level >= 0 {
		curr = pred.next[level].Load()

		for key > curr.key {
			pred = curr
			curr = pred.next[level].Load()
		}

		//fmt.Println(key)
		//fmt.Println(curr.key)

		if foundLevel == -1 && key == curr.key {
			foundLevel = level
		}

		preds[level] = pred
		succs[level] = curr
		level = level - 1
	}
	slog.Info("returning from find help")

	return foundLevel, preds, succs
}

// Takes in a key and tries to find that node in the skiplist
// Returns the node and true on success, nil and false on failure
func (s *List[K, V]) Find(key K) (V, bool) {

	levelFound, _, succs := s.findHelp(key)

	if levelFound == -1 {
		slog.Info("object not found")
		return *new(V), false
	}
	slog.Info("object found")

	found := succs[levelFound]

	return found.value, found.fullyLinked.Load() && !found.marked.Load()
}

// From slides, Insert algorithm (need to modify for updating too later)
// Takes in the key and updatecheck function and tries to insert that node.
// Returns true as its first value if the node was updated, and false otherwise. It returns an error if encountered as its second value
func (s *List[K, V]) Upsert(key K, check UpdateCheck[K, V]) (updated bool, err error) {

	// we don't want to insert an empty string:
	if key <= s.head.key {
		return false, fmt.Errorf("trying to insert an empty string")
	}

	// Pick random top level, .5 chance of being 0, .25 of being 1, e.t.c
	randnum := rand.Intn(100) // this is 0 to 99

	// topLevel has 50% change of being 0, 25% chance of being 1, e.t.c.
	var topLevel int
	if randnum >= 50 {
		topLevel = 0
	} else if randnum >= 25 {
		topLevel = 1
	} else if randnum >= 12 {
		topLevel = 2
	} else {
		topLevel = 3
	}

	slog.Info("After toplevel")

	// Keep trying to insert until success or failure
	for {
		slog.Info("line 224")
		levelFound, preds, succs := s.findHelp(key)
		var found *node[K, V]

		if levelFound != -1 {
			found = succs[levelFound]

			if !found.marked.Load() {
				// Node is being added, wait for that to finish
				for !found.fullyLinked.Load() {
					// do nothing
				}
				// this is the case of updating since you found the node
				val, err := check(key, found.value, true)
				if err != nil {
					slog.Info("Error != nil in upsert")
					found.value = val
					return true, err
				}
				slog.Info("error == nil in upsert")
				// is this enough to update the value?
				found.value = val
				return true, err
			}
			// otherwise, found node is currently being removed, try again
			continue
		}
		slog.Info("After first if")

		// Moving to Slide 18
		highestLocked := -1
		valid := true
		level := 0

		// keep a map of node to level, then if its in the map you know you already locked it
		lockednodes := make(map[*node[K, V]]bool)

		// Lock all predecessors
		for valid && level <= topLevel {
			slog.Info("locking")
			// gotta make sure you haven't locked this yet. also note that this node is locked
			if preds[level].mtx.TryLock() {
				lockednodes[preds[level]] = true
			}

			highestLocked = level
			slog.Info("finished lock")

			// Make sure pred and succ are still valid
			unmarked := !preds[level].marked.Load() && !succs[level].marked.Load()
			connected := preds[level].next[level].Load() == succs[level]
			valid = unmarked && connected

			// move to the next highest level
			level = level + 1
		}
		slog.Info("After lock all preds")

		if !valid {
			// Preds or succs changed, unlock and try again
			level = highestLocked
			slog.Info("node was either marked for removal or not fully connected")
			for level >= 0 {
				if lockednodes[preds[level]] {
					preds[level].mtx.Unlock()
					lockednodes[preds[level]] = false
				}
				level = level - 1
			}
			continue
		}
		slog.Info("After !valid")

		// this is the case of updating
		if levelFound != -1 {
			slog.Info("INSIDE if levelFound != -1")
			val, err := check(key, found.value, true)
			if err != nil {
				slog.Info("Error != nil in upsert")
				return false, err
			}
			slog.Info("error == nil in upsert")
			// is this enough to update the value
			found.value = val
			return true, err
		}
		slog.Info("After case for updating")

		// this is the case of inserting a new node

		var currVal V
		val, err := check(key, currVal, false)
		if err != nil {
			slog.Info("error: sending false to check function and this print says that it sent true")
		}

		var newnode node[K, V] = node[K, V]{
			next:     make([]atomic.Pointer[node[K, V]], s.maxlevel+1), // + 1 ??
			key:      key,
			value:    val,
			topLevel: topLevel,
			//marked: atomic.Bool{}, false
			//fullyLinked: atomic.Bool{}, false
		}

		// Set next pointers
		level = 0
		for level <= topLevel {
			newnode.next[level].Store(succs[level])
			level = level + 1
		}
		slog.Info("after setting pointers")

		// Add to skip list from bottom
		level = 0
		for level <= topLevel {
			preds[level].next[level].Store(&newnode)
			level = level + 1
		}
		slog.Info("after adding to skip list")

		// Node has been added
		newnode.fullyLinked.Store(true)

		slog.Info("after storing node")

		// Unlocking the keys of the map of lockednodes
		for node := range lockednodes {
			node.mtx.Unlock()
		}

		slog.Info("after unlocking")

		// Increment timestamp for Query function
		s.timestamp.Add(1)

		slog.Info("end of upsert")

		return false, nil
	}
}

// Takes in a key and attempts to remove that node from the skiplist.
// Returns the removed value and true on success, or an empty value and false on failure.
func (s *List[K, V]) Remove(key K) (removedValue V, removed bool) {
	var victim *node[K, V] // Victim node to remove
	isMarked := false      // Have we already marked the victim?
	topLevel := -1         // Top level of victim node

	for {
		slog.Info("line 412")
		levelFound, preds, succs := s.findHelp(key)
		if levelFound == -1 {
			slog.Info("did not find it to be removed")
		}
		if levelFound != -1 {
			slog.Info("found it in remove")
			victim = succs[levelFound]
		}
		slog.Info("before !isMarked")
		var emptyVal V
		if !isMarked {
			slog.Info("in !isMarked")
			// First time through
			if levelFound == -1 {
				slog.Info("it is not found")
				// No matching node found
				// return <nothing>, false. nil does not work?
				return emptyVal, false
			}

			if !victim.fullyLinked.Load() {
				slog.Info("it is not fullyLinked")
				// Victim not yet inserted
				return emptyVal, false
			}

			if victim.marked.Load() {
				slog.Info("it is marked")
				// Victim already being removed
				return emptyVal, false
			}

			if victim.topLevel != levelFound {
				slog.Info("the victim's toplevel is not the found level?")
				// Wasn't fullyLinked when found
				return emptyVal, false
			}

			topLevel = victim.topLevel
			slog.Info("locking the victim")
			victim.mtx.Lock()
			if victim.marked.Load() {
				slog.Info("unlocking the victim, it is marked")
				// Another remove call beat us
				victim.mtx.Unlock()
				//return victim.value, false
				return emptyVal, false
			}

			slog.Info("marking the victim")
			victim.marked.Store(true)
			isMarked = true

		}
		slog.Info("after !isMarked")

		// Victim is locked and marked
		highestLocked := -1
		level := 0
		valid := true

		var pred *node[K, V]
		var succCheck bool

		// keep a map of node to level, if its in the map you know you already locked it
		lockednodes := make(map[*node[K, V]]bool)

		for valid && (level <= topLevel) {
			pred = preds[level]
			if pred.mtx.TryLock() {
				lockednodes[preds[level]] = true
			}
			highestLocked = level
			succCheck = (pred.next[level].Load() == victim)
			valid = !pred.marked.Load() && succCheck
			level = level + 1
		}

		slog.Info("before !valid")

		if !valid {
			// Unlock
			level = highestLocked
			for level >= 0 {
				if lockednodes[preds[level]] {
					preds[level].mtx.Unlock()
					lockednodes[preds[level]] = false
				}
				level = level - 1
			}
			// Predecessors changed, try again
			// victim remains locked and marked
			continue
		}

		slog.Info("after !valid")

		// All preds are locked and valid, unlink
		level = topLevel
		for level >= 0 {
			// Not sure about this line
			preds[level].next[level].Store(victim.next[level].Load())
			level = level - 1
		}
		// Unlock
		victim.mtx.Unlock()
		level = highestLocked
		for level >= 0 {
			if lockednodes[preds[level]] {
				preds[level].mtx.Unlock()
				lockednodes[preds[level]] = false
			}
			level = level - 1
		}

		// Increment timestamp for Query function
		s.timestamp.Add(1)

		slog.Info("returning from Remove")

		return victim.value, true
	}
}

// Takes in a start and end parameter that defines what range to query on.
// Returns a slice of key value pairs that represent the nodes found.
func (s *List[K, V]) Query(start K, end K) (results []Pair[K, V]) {
	slog.Info("In Query")
	startString := fmt.Sprint(start)
	endString := fmt.Sprint(end)
	if startString == "" {
		startString = "a"
	}
	if endString == "" {
		endString = fmt.Sprint(s.tail.key)
		endString = endString[:len(endString)-1] + "y"
	}
	slog.Info("start string", startString)
	slog.Info("end string", endString)
	var curString string
	for {
		// Initialize a condition to tryAgain, and a timestamp of first check
		curr := s.head
		curString = fmt.Sprint(curr.key)

		//savedList := make([]atomic.Pointer[node[K, V]],  what would length be?)
		var savedList []Pair[K, V]
		var p Pair[K, V]

		stmp1 := s.timestamp.Load()

		for curString <= endString {
			// iterate through the list at level 0, saving each node in the list

			slog.Info("current docName ", curString)
			if curString >= startString {
				p = Pair[K, V]{Key: curr.key, Value: curr.value}
				savedList = append(savedList, p)
			}
			curr = curr.next[0].Load()
			curString = fmt.Sprint(curr.key)

		}

		// do a second check of the list, to see if it remained the same
		stmp2 := s.timestamp.Load()

		if stmp1 != stmp2 {
			continue
		}
		return savedList
	}
}
