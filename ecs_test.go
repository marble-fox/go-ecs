package go_ecs

import (
	"errors"
	"log"
	"testing"
)

type Position struct {
	X, Y float64
}

type Velocity struct {
	X, Y float64
}

type Scale struct {
	X, Y float64
}

var world1 *World
var world2 *World

var posStorage1 *ComponentStorage[*Position]
var velStorage1 *ComponentStorage[*Velocity]
var scaleStorage1 *ComponentStorage[*Scale]

var posStorage2 *ComponentStorage[*Position]

var e1w1 EntityID
var e2w1 EntityID
var e1w2 EntityID

// TODO improve tests
func TestEntityID(t *testing.T) {
	testID := makeEntityID(123, 1337, 69)
	if testID.Index() != 123 {
		t.Errorf("Index is wrong: %v", testID.Index())
	}
	if testID.Generation() != 1337 {
		t.Errorf("Generation is wrong: %v", testID.Generation())
	}
	if testID.WorldID() != 69 {
		t.Errorf("WorldID is wrong: %v", testID.WorldID())
	}

	log.Printf("EntityID generation tested")
}

func TestECSBasic(t *testing.T) {
	var err error

	// Create World
	world1, err = CreateWorld("world1")
	if err != nil {
		t.Fatalf("failed to create world1: %v", err)
	}

	world2, err = CreateWorld("world2")
	if err != nil {
		t.Fatalf("failed to create world2: %v", err)
	}

	// Create Component Storages for world1
	posStorage1, err = CreateComponent[*Position](world1)
	if err != nil {
		t.Fatalf("failed to create position storage: %v", err)
	}
	velStorage1, err = CreateComponent[*Velocity](world1)
	if err != nil {
		t.Fatalf("failed to create velocity storage: %v", err)
	}
	scaleStorage1, err = CreateComponent[*Scale](world1)
	if err != nil {
		t.Fatalf("failed to create scale storage: %v", err)
	}

	// Create Component Storages for world2
	posStorage2, err = CreateComponent[*Position](world2)
	if err != nil {
		t.Fatalf("failed to create position storage: %v", err)
	}

	// Create Entities in world1
	e1w1, err = world1.CreateEntity()
	if err != nil {
		t.Fatalf("failed to create entity 1 for world1: %v", err)
	}
	e2w1, err = world1.CreateEntity()
	if err != nil {
		t.Fatalf("failed to create entity 2 for world1: %v", err)
	}

	// Create Entities in world2
	e1w2, err = world2.CreateEntity()
	if err != nil {
		t.Fatalf("failed to create entity 1 for world2: %v", err)
	}

	// Register Components for world1 entities
	pos_e1ps1 := &Position{X: 10, Y: 20}
	vel_e1vs1 := &Velocity{X: 1, Y: 2}
	scale_e1ss1 := &Scale{X: 1.5, Y: 1.5}
	err = posStorage1.Register(e1w1, pos_e1ps1)
	if err != nil {
		t.Fatalf("failed to register position for e1w1: %v", err)
	}
	err = velStorage1.Register(e1w1, vel_e1vs1)
	if err != nil {
		t.Fatalf("failed to register velocity for e1w1: %v", err)
	}
	err = scaleStorage1.Register(e1w1, scale_e1ss1)
	if err != nil {
		t.Fatalf("failed to register scale for e1w1: %v", err)
	}

	pos_e2ps1 := &Position{X: 100, Y: 200}
	err = posStorage1.Register(e2w1, pos_e2ps1)
	if err != nil {
		t.Fatalf("failed to register position for e2w1: %v", err)
	}

	// Register Components for world2 entities
	pos_e1ps2 := &Position{X: -10, Y: -20}
	err = posStorage2.Register(e1w2, pos_e1ps2)
	if err != nil {
		t.Fatalf("failed to register position for e1w2: %v", err)
	}

	// Register wrong world entity
	err = posStorage2.Register(e1w1, pos_e1ps1)
	if err == nil || !errors.Is(err, ErrWrongWorldID) {
		t.Fatalf("expected error registering position for wrong world entity")
	}

	//Each iteration over world1 positions
	count1 := 0
	err = posStorage1.Each(func(id EntityID, p *Position) error {
		count1++
		return nil
	})
	if err != nil {
		t.Fatalf("failed to iterate over world1 posStorage Each: %v", err)
	}
	if count1 != 2 {
		t.Errorf("expected 2 entities in world1 posStorage Each, got %d", count1)
	}

	// Verify world1 data
	getPos_e1ps1, _ := posStorage1.Get(e1w1)
	if getPos_e1ps1.X != 10 || getPos_e1ps1.Y != 20 {
		t.Fatalf("world1 e1w1: expected position {10, 20}, got %+v", getPos_e1ps1)
	}
	getVel_e1vs1, _ := velStorage1.Get(e1w1)
	if getVel_e1vs1.X != 1 || getVel_e1vs1.Y != 2 {
		t.Fatalf("world1 e1w1: expected velocity {1, 2}, got %+v", getVel_e1vs1)
	}
	getScale_e1ss1, _ := scaleStorage1.Get(e1w1)
	if getScale_e1ss1.X != 1.5 || getScale_e1ss1.Y != 1.5 {
		t.Fatalf("world1 e1w1: expected scale {1.5, 1.5}, got %+v", getScale_e1ss1)
	}

	getPos_e2ps1, _ := posStorage1.Get(e2w1)
	if getPos_e2ps1.X != 100 || getPos_e2ps1.Y != 200 {
		t.Fatalf("world1 e2w1: expected position {100, 200}, got %+v", getPos_e2ps1)
	}

	// Verify world2 data
	getPos_e1ps2, _ := posStorage2.Get(e1w2)
	if getPos_e1ps2.X != -10 || getPos_e1ps2.Y != -20 {
		t.Fatalf("world2 e1w2: expected position {-10, -20}, got %+v", getPos_e1ps2)
	}

	// Verify isolation: e1w1 should not have a position in world2's posStorage
	if posStorage2.IsEntityIDRegistered(e1w1) {
		t.Fatalf("world1 entity e1w1 should not be registered in world2's position storage")
	}

	// Delete Entity in world1
	err = world1.DeleteEntity(e1w1)
	if err != nil {
		t.Fatalf("failed to delete entity 1 for world1: %v", err)
	}
	if world1.IsEntityAlive(e1w1) {
		t.Fatalf("e1w1 should not be alive in world1 after deletion")
	}
	if posStorage1.IsEntityIDRegistered(e1w1) {
		t.Fatalf("e1w1 position should be removed from world1 posStorage after deletion")
	}
	if velStorage1.IsEntityIDRegistered(e1w1) {
		t.Fatalf("e1w1 velocity should be removed from world1 velStorage after deletion")
	}

	// Ensure world2 entity is still alive
	if !world2.IsEntityAlive(e1w2) {
		t.Fatalf("e1w2 in world2 should still be alive")
	}

	// Test Update and Get
	newPos := &Position{X: 50, Y: 60}
	err = posStorage1.Update(e2w1, newPos)
	if err != nil {
		t.Fatalf("failed to update position for e2w1: %v", err)
	}
	updatedPos, err := posStorage1.Get(e2w1)
	if err != nil || updatedPos.X != 50 || updatedPos.Y != 60 {
		t.Fatalf("expected updated position {50, 60}, got %+v (err: %v)", updatedPos, err)
	}

	// Test GetEntitiesIDs
	//w1Entities := world1.GetEntitiesIDs()
	//if len(w1Entities) != 1 || w1Entities[0] != e2w1 {
	//	t.Fatalf("expected 1 entity (e2w1) in world1, got %v", w1Entities)
	//}

	// Test Entity ID reuse and generation
	e3w1, err := world1.CreateEntity()
	if err != nil {
		t.Fatalf("failed to create e3w1: %v", err)
	}
	if e3w1.Index() != e1w1.Index() {
		t.Fatalf("expected reused index %d, got %d", e1w1.Index(), e3w1.Index())
	}
	if e3w1.Generation() <= e1w1.Generation() {
		t.Fatalf("expected generation to increase: e1w1 gen %d, e3w1 gen %d", e1w1.Generation(), e3w1.Generation())
	}

	// Test Error: Register existing component
	err = posStorage1.Register(e2w1, &Position{})
	if !errors.Is(err, ErrEntityIDAlreadyRegistered) {
		t.Fatalf("expected ErrEntityIDAlreadyRegistered, got %v", err)
	}

	// Test Error: Operation on deleted entity
	_, err = posStorage1.Get(e1w1)
	if !errors.Is(err, ErrEntityNotAlive) {
		t.Fatalf("expected ErrEntityNotAlive for deleted entity, got %v", err)
	}

	// Test GetWorldByName
	w1Again, err := GetWorldByName("world1")
	if err != nil || w1Again != world1 {
		t.Fatalf("failed to get world1 by name: %v", err)
	}

	// Test Error: Create world with duplicate name
	_, err = CreateWorld("world1")
	if !errors.Is(err, ErrWorldAlreadyExists) {
		t.Fatalf("expected ErrWorldAlreadyExists, got %v", err)
	}

	// Final each's
	err = posStorage1.Each(func(id EntityID, p *Position) error {
		t.Logf("posStorage1: %v, %v\n", id, p)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to iterate over posStorage1: %v", err)
	}

	err = velStorage1.Each(func(id EntityID, v *Velocity) error {
		t.Logf("velStorage1: %v, %v\n", id, v)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to iterate over velStorage1: %v", err)
	}

	err = scaleStorage1.Each(func(id EntityID, s *Scale) error {
		t.Logf("scaleStorage1: %v, %v\n", id, s)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to iterate over scaleStorage1: %v", err)
	}

	err = posStorage2.Each(func(id EntityID, p *Position) error {
		t.Logf("posStorage2: %v, %v\n", id, p)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to iterate over posStorage2: %v", err)
	}

	// Test DestroyWorldByName
	err = DestroyWorldByName("world1")
	if err != nil {
		t.Fatalf("failed to destroy world1 by name: %v", err)
	}
	if world1.IsAlive() {
		t.Fatalf("world1 should be destroyed")
	}

	// Test Error: Operation on destroyed world
	_, err = world1.CreateEntity()
	if !errors.Is(err, ErrWorldIsNotAlive) {
		t.Fatalf("expected ErrWorldIsNotAlive, got %v", err)
	}

	// Delete world2
	err = world2.Destroy()
	if err != nil {
		t.Fatalf("failed to destroy world2: %v", err)
	}

	// Test GetWorldByName after destruction
	_, err = GetWorldByName("notAWorld")
	if !errors.Is(err, ErrWorldDoesntExist) {
		t.Fatalf("expected ErrWorldDoesntExist for notAWorld, got %v", err)
	}

	// Test GetWorldByName after destruction
	_, err = GetWorldByName("world2")
	if !errors.Is(err, ErrWorldDoesntExist) {
		t.Fatalf("expected ErrWorldDoesntExist for world2, got %v", err)
	}
}
