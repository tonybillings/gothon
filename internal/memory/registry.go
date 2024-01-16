package memory

type Registry map[string]Register

func (r Registry) Init() {
	for _, reg := range r {
		reg.Init()
	}
}

func NewRegistry(registers []Register) Registry {
	registry := make(map[string]Register)

	for _, r := range registers {
		registry[r.ID()] = r
	}

	return registry
}
