package go_ecs

import "errors"

// ErrPageAlreadyExists is returned if a sparse page is unexpectedly re-created.
var ErrPageAlreadyExists = errors.New("sparse page already exists")

// calculatePageIndexAndDenseIndexOnPage calculates the page index and the position within that page for an sparseID.
func calculatePageIndexAndDenseIndexOnPage(entityIndex uint32) (int, int) {
	index := int(entityIndex)

	sparsePageIndex := index / maxPageSize
	entityIndexOnPage := index % maxPageSize

	return sparsePageIndex, entityIndexOnPage
}

// addNewValueToDense appends a new value to the dense array and returns its index.
func (ss *sparseSet[T]) addNewValueToDense(entityID EntityID, value T) uint32 {
	valueIndex := ss.denseSize
	ss.dense = append(ss.dense, value)
	ss.denseToSparse = append(ss.denseToSparse, entityID.Index())
	ss.denseSize++

	return valueIndex
}

// createNewSparsePage creates a new sparse page and returns a pointer to it.
func (ss *sparseSet[T]) createNewSparsePage(sparsePageIndex int) *sparsePage {
	if len(ss.sparseBook) <= sparsePageIndex {
		// Fill sparseBook with empty pages
		pages := make([]*sparsePage, sparsePageIndex-len(ss.sparseBook)+1)
		ss.sparseBook = append(ss.sparseBook, pages...)
	}

	if ss.sparseBook[sparsePageIndex] != nil {
		panic(ErrPageAlreadyExists)
	}

	var newPage sparsePage
	page := &newPage
	ss.sparseBook[sparsePageIndex] = page

	return page
}
