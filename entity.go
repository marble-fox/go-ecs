package go_ecs

import (
	"errors"
)

// ErrEntityNotAlive is returned when an operation is attempted on an entity that does not exist or has been deleted.
var ErrEntityNotAlive = errors.New("entity does is not alive")

// ErrEntityIDIsZero is returned when an entity ID is unexpectedly 0.
var ErrEntityIDIsZero = errors.New("entityID cannot be 0")

// ErrEntityIDAlreadyRegistered is returned when an entity ID is already registered in a component storage.
var ErrEntityIDAlreadyRegistered = errors.New("entityID already registered")

// ErrEntityIDNotFound is returned when an entity is not found in a component storage.
var ErrEntityIDNotFound = errors.New("entityID not found in the book")

// ErrWrongWorldID is returned when an entity ID belongs to a different world.
var ErrWrongWorldID = errors.New("lastWorldID does not match")

// EntityID represents a unique identifier for an entity in the ECS.
// It combines a 16-bit world ID, a 16-bit generation, and a 32-bit index.
type EntityID uint64

// entitySlot stores the state of an entity index in the world.
type entitySlot struct {
	generation uint16 // Increases every time the slot is reused to invalidate old EntityIDs
	alive      bool   // True if the entity is currently active
}

// makeEntityID creates a new EntityID from an index, a generation, and a lastWorldID.
func makeEntityID(index uint32, generation uint16, worldID uint16) EntityID {
	return EntityID(uint64(worldID)<<48 | uint64(generation)<<32 | uint64(index))
}

// Index returns the 32-bit index part of the EntityID.
func (entityID EntityID) Index() uint32 {
	return uint32(entityID)
}

// Generation returns the 16-bit generation part of the EntityID.
func (entityID EntityID) Generation() uint16 {
	return uint16((entityID >> 32) & 0xFFFF)
}

// WorldID returns the 16-bit world ID part of the EntityID.
func (entityID EntityID) WorldID() uint16 {
	return uint16(entityID >> 48)
}

// CreateEntity creates a new entity in the world and returns its ID.
func (world *World) CreateEntity() (EntityID, error) {
	if world.alive == false {
		return 0, ErrWorldIsNotAlive
	}

	if len(world.freeEntityIndices) > 0 {
		// Reuse existing entity slot
		firstFreeEntitySlotIndex := world.freeEntityIndices[0]
		world.freeEntityIndices = world.freeEntityIndices[1:]

		slot := &world.entitySlots[firstFreeEntitySlotIndex]
		slot.generation++
		slot.alive = true

		newEntityID := makeEntityID(firstFreeEntitySlotIndex, slot.generation, world.ID)
		return newEntityID, nil
	}

	// New entity slot
	newEntitySlotIndex := len(world.entitySlots)
	slot := entitySlot{0, true}
	world.entitySlots = append(world.entitySlots, slot)

	newEntityID := makeEntityID(uint32(newEntitySlotIndex), 0, world.ID)
	return newEntityID, nil
}

// DeleteEntity marks an entity as deleted in the world and cleans up associated resources.
func (world *World) DeleteEntity(entityID EntityID) error {
	if entityID.WorldID() != world.ID {
		return ErrWrongWorldID
	}

	if !world.IsEntityAlive(entityID) {
		return ErrEntityNotAlive
	}

	// Remove the entity from all component storages
	var deleteErrors []error
	for _, storage := range world.componentStorages {
		if !storage.IsEntityIDRegistered(entityID) {
			continue
		}

		err := storage.deleteRegisteredEntity(entityID)
		if err != nil {
			deleteErrors = append(deleteErrors, err)
		}
	}

	// Mark the entity slot as dead
	entitySlotIndex := entityID.Index()
	slot := &world.entitySlots[entitySlotIndex]
	slot.alive = false

	// Allow reuse of this slot if its generation variable is not full
	if slot.generation < 65535 {
		slot.generation++
		world.freeEntityIndices = append(world.freeEntityIndices, entitySlotIndex)
	}

	if len(deleteErrors) > 0 {
		return errors.Join(deleteErrors...)
	}

	return nil
}

// IsEntityAlive checks if an entity is alive in the world.
func (world *World) IsEntityAlive(entityID EntityID) bool {
	entitySlotIndex := entityID.Index()

	// Check if the entity slot exists
	if len(world.entitySlots) < int(entitySlotIndex) {
		return false
	}

	// Check is generation matches
	slot := &world.entitySlots[entitySlotIndex]
	if entityID.Generation() != slot.generation {
		return false
	}

	return slot.alive
}
