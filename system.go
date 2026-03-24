package go_ecs

const maxSystemPriority uint8 = 255

func (world *World) AddSystem(system func(world *World) error, priority uint8) {
	world.systems[priority] = append(world.systems[priority], system)
}

// RunSystems executes all registered systems in the world in ascending order by priority.
// Returns the first error encountered during system execution, if any.
func (world *World) RunSystems() error {
	for _, systemSlice := range world.systems {
		for _, system := range systemSlice {
			err := system(world)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
