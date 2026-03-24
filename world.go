package go_ecs

import (
	"errors"
)

// ErrWorldAlreadyExists is returned when trying to create a world with a name that is already in use.
var ErrWorldAlreadyExists = errors.New("world with this name already exists")

// ErrWorldDoesntExist is returned when looking up a world by name that hasn't been created.
var ErrWorldDoesntExist = errors.New("world with this name does not exist")

// ErrWorldIsNotAlive is returned when performing operations on a world that has been destroyed.
var ErrWorldIsNotAlive = errors.New("world is not alive")

// TODO make generation system for worlds
var lastWorldID uint16 = 1

// World represents a container for entities and their components.
// It manages entity lifecycles and provides access to component storages.
type World struct {
	ID                uint16                // Unique ID for the world
	Name              string                // Unique name of the world
	Active            bool                  // Whether the world is currently active
	alive             bool                  // Internal flag indicating if the world is not destroyed
	componentStorages []AnyComponentStorage // List of all component storages registered in this world
	entitySlots       []entitySlot          // Internal list of all entity slots (alive and dead)
	freeEntityIndices []uint32              // Indices of dead entity slots that can be reused
	systems           [maxSystemPriority][]func(world *World) error
}

// worlds is a global registry of all created worlds by their name.
// TODO make it private
var worlds map[string]*World = make(map[string]*World)

// GetWorlds TODO commenting
func GetWorlds() map[string]*World {
	return worlds
}

// CreateWorld initializes a new world with the given name and adds it to the global registry.
// If a world with the same name already exists, it returns ErrWorldAlreadyExists.
func CreateWorld(newWorldName string) (*World, error) {
	_, ok := worlds[newWorldName]
	if ok {
		return nil, ErrWorldAlreadyExists
	}

	newWorldID := lastWorldID
	lastWorldID++

	newWorld := &World{
		ID:                newWorldID,
		Name:              newWorldName,
		Active:            true,
		alive:             true,
		componentStorages: make([]AnyComponentStorage, 0),
		entitySlots:       make([]entitySlot, 0),
		freeEntityIndices: make([]uint32, 0),
		systems:           [maxSystemPriority][]func(world *World) error{},
	}
	worlds[newWorldName] = newWorld

	return newWorld, nil
}

// IsAlive returns true if the world has not been destroyed.
func (world *World) IsAlive() bool {
	return world.alive
}

// Destroy removes the world from the registry and deletes all its entities and components.
// Returns an error if the world is not alive.
func (world *World) Destroy() error {
	world.alive = false
	delete(worlds, world.Name)

	for slotIndex, slot := range world.entitySlots {
		if !slot.alive {
			continue
		}

		entityID := makeEntityID(uint32(slotIndex), slot.generation, world.ID)
		err := world.DeleteEntity(entityID)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetWorldByName retrieves a world from the global registry by its name.
func GetWorldByName(name string) (*World, error) {
	world, ok := worlds[name]
	if !ok {
		return nil, ErrWorldDoesntExist
	}

	return world, nil
}

// DestroyWorldByName looks up a world by name and destroys it if found.
func DestroyWorldByName(name string) error {
	world, err := GetWorldByName(name)
	if err != nil {
		return err
	}

	err = world.Destroy()
	if err != nil {
		return err
	}
	return nil
}

// GetComponentStorage returns the component storage of the specified type T from the world.
// It returns nil if the storage for type T is not found.
func GetComponentStorage[T any](world *World) (*ComponentStorage[T], bool) {
	for _, storage := range world.componentStorages {
		componentStorage, ok := storage.(*ComponentStorage[T])
		if ok {
			return componentStorage, true
		}
	}
	return nil, false
}
