package go_ecs

import "errors"

var ErrPageDoesntExits = errors.New("sparse page doesn't exist")
var ErrIndexNotFound = errors.New("index not found")

// maxPageSize defines the number of entities handled by a single sparse page.
const maxPageSize int = 256

// sparsePage stores indices into the dense array.
type sparsePage [maxPageSize]uint32

// sparseSet is a sparse set implementation that stores values in a dense array.
type sparseSet[T any] struct {
	sparseBook    []*sparsePage // Slice of sparse pages for faster lookup
	dense         []T           // Contiguous array of component values
	denseToSparse []uint32      // Maps dense index back to sparse book
	denseSize     uint32        // Number of elements in the dense array (including the 0-th dummy element)
}

// NewSparseSet creates and initializes a new sparseSet.
func newSparseSet[T any]() sparseSet[T] {
	return sparseSet[T]{
		sparseBook:    make([]*sparsePage, 0),
		dense:         make([]T, 1),      // Start with size 1 to use 0 as "not found"
		denseToSparse: make([]uint32, 1), // Start with size 1 to use 0 as "not found"
		denseSize:     1,
	}
}

// set TODO commenting
func (ss *sparseSet[T]) set(entityID EntityID, value T) {
	sparsePageIndex, denseIndexOnPage := calculatePageIndexAndDenseIndexOnPage(entityID.Index())

	var page *sparsePage
	if len(ss.sparseBook) <= sparsePageIndex || ss.sparseBook[sparsePageIndex] == nil {
		page = ss.createNewSparsePage(sparsePageIndex)
	} else {
		page = ss.sparseBook[sparsePageIndex]
	}

	indexToDense := page[denseIndexOnPage]
	if indexToDense != 0 {
		ss.dense[indexToDense] = value
		return
	}

	// NewSparseSet value
	indexToDense = ss.addNewValueToDense(entityID, value)
	page[denseIndexOnPage] = indexToDense
}

// get TODO commenting
func (ss *sparseSet[T]) get(entityID EntityID) (T, bool) {
	sparsePageIndex, denseIndexOnPage := calculatePageIndexAndDenseIndexOnPage(entityID.Index())

	if len(ss.sparseBook) <= sparsePageIndex || ss.sparseBook[sparsePageIndex] == nil {
		var zero T
		return zero, false
	}

	page := ss.sparseBook[sparsePageIndex]
	denseIndex := page[denseIndexOnPage]

	if denseIndex == 0 {
		var zero T
		return zero, false
	}

	return ss.dense[denseIndex], true
}

// delete TODO commenting
func (ss *sparseSet[T]) delete(entityID EntityID) error {
	sparsePageIndex, denseIndexOnPage := calculatePageIndexAndDenseIndexOnPage(entityID.Index())

	if len(ss.sparseBook) <= sparsePageIndex || ss.sparseBook[sparsePageIndex] == nil {
		return ErrPageDoesntExits
	}

	denseIndex := ss.sparseBook[sparsePageIndex][denseIndexOnPage]
	if denseIndex == 0 {
		return ErrIndexNotFound
	}

	// Remove index to dense
	ss.sparseBook[sparsePageIndex][denseIndexOnPage] = 0

	indexToLastValue := ss.denseSize - 1
	if denseIndex != indexToLastValue {
		// Swap deleted element with the last element
		entityIDOfTheLastValue := ss.denseToSparse[indexToLastValue]
		lastValueEntityIDPageIndex, lastValueEntityIDOnPage := calculatePageIndexAndDenseIndexOnPage(entityIDOfTheLastValue)

		ss.sparseBook[lastValueEntityIDPageIndex][lastValueEntityIDOnPage] = denseIndex

		ss.dense[denseIndex] = ss.dense[indexToLastValue]
		ss.denseToSparse[denseIndex] = ss.denseToSparse[indexToLastValue]
	}

	// Shrink arrays
	ss.dense = ss.dense[:indexToLastValue]
	ss.denseToSparse = ss.denseToSparse[:indexToLastValue]
	ss.denseSize--

	return nil
}
