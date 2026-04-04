package go_ecs

import (
	"errors"
)

// ErrComponentAlreadyExists is returned when trying to register a component that already exists.
var ErrComponentAlreadyExists = errors.New("component already exists")

// ErrComponentNotRegistered is returned when trying to access a component that is not registered.
var ErrComponentNotRegistered = errors.New("component not registered")

// AnyComponentStorage is an interface for component storages regardless of the component type.
// It allows for generic operations like removing an entity from all storages.
type AnyComponentStorage interface {
	IsEntityIDRegistered(entityID EntityID) bool
	deleteRegisteredEntity(entityID EntityID) error
}

// ComponentStorage manages components of a specific type T for entities in a world.
// It uses a sparse set for efficient storage and O(1) access.
type ComponentStorage[T any] struct {
	World *World
	data  *sparseSet[T]
}

// registerComponentStorage adds a storage to a world's list of component storages.
func registerComponentStorage(world *World, storage AnyComponentStorage) {
	world.componentStorages = append(world.componentStorages, storage)
}

// CreateComponent creates and registers a new component storage for type T in the given world.
// Returns an error if the world is not alive or if the component storage already exists.
func CreateComponent[T any](world *World) (*ComponentStorage[T], error) {
	if !world.alive {
		return nil, ErrWorldIsNotAlive
	}

	_, ok := GetComponentStorage[T](world)
	if ok {
		return nil, ErrComponentAlreadyExists
	}

	storage := &ComponentStorage[T]{
		World: world,
		data:  new(newSparseSet[T]()),
	}
	registerComponentStorage(world, storage)

	return storage, nil
}

// Register associates a component value with an entity in the storage.
// It returns an error if the entity is not alive or already has a component in this storage.
func (storage *ComponentStorage[T]) Register(entityID EntityID, value T) error {
	if storage.World.ID != entityID.WorldID() {
		return ErrWrongWorldID
	}

	err := validateEntityID[T](entityID, *storage)
	if err != nil {
		return err
	}

	_, ok := storage.data.get(entityID)
	if ok {
		return ErrEntityIDAlreadyRegistered
	}

	storage.data.set(entityID, value)

	return nil
}

// Update changes the component value associated with an entity.
func (storage *ComponentStorage[T]) Update(entityID EntityID, value T) error {
	err := validateEntityID[T](entityID, *storage)
	if err != nil {
		return err
	}

	storage.data.set(entityID, value)

	return nil
}

// Get retrieves the component value associated with an entity.
// Returns an error if the entity is not alive or not registered in this storage.
func (storage *ComponentStorage[T]) Get(entityID EntityID) (T, error) {
	err := validateEntityID[T](entityID, *storage)
	if err != nil {
		return *new(T), err
	}

	value, ok := storage.data.get(entityID)
	if !ok {
		return *new(T), ErrComponentNotRegistered
	}

	return value, nil
}

// deleteRegisteredEntity removes the component associated with the entity from the storage.
func (storage *ComponentStorage[T]) deleteRegisteredEntity(entityID EntityID) error {
	return storage.data.delete(entityID)
}

// IsEntityIDRegistered checks if the entity has a component in this storage.
func (storage *ComponentStorage[T]) IsEntityIDRegistered(entityID EntityID) bool {
	if storage.World.ID != entityID.WorldID() {
		return false
	}

	_, ok := storage.data.get(entityID)
	return ok
}

// Each iterates over all entities in the storage and calls the provided function for each one.
func (storage *ComponentStorage[T]) Each(fn func(EntityID, T) error) error {
	storageWorldSlots := storage.World.entitySlots

	// Copy the data to avoid changing the slice while iterating over it
	denseCopy := storage.data.dense
	denseToSparseCopy := storage.data.denseToSparse

	errorsSlice := make([]error, 0)
	for i, value := range denseCopy {
		if i == 0 {
			continue
		}

		entityIndex := denseToSparseCopy[i]
		entityID := makeEntityID(entityIndex, storageWorldSlots[entityIndex].generation, storage.World.ID)

		err := fn(entityID, value)
		if err != nil {
			errorsSlice = append(errorsSlice, err)
		}
	}

	if len(errorsSlice) > 0 {
		return errors.Join(errorsSlice...)
	}

	return nil
}

// TODO parallel each

// validateEntityID checks if an EntityID is valid for operations in the given storage.
func validateEntityID[T any](entityID EntityID, storage ComponentStorage[T]) error {
	if storage.World.ID != entityID.WorldID() {
		return ErrWrongWorldID
	}

	if !storage.World.alive {
		return ErrWorldIsNotAlive
	}

	if !storage.World.IsEntityAlive(entityID) {
		return ErrEntityNotAlive
	}

	return nil
}
