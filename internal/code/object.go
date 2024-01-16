package code

type VariableType byte

const (
	Bool VariableType = iota
	Int
	Float
	Str
	LockFunc
	UnlockFunc
	WaitGroup
	Queue
	LifoQueue
)

func (v VariableType) String() string {
	switch v {
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Float:
		return "float"
	case Str:
		return "str"
	case LockFunc:
		return "lock_func"
	case UnlockFunc:
		return "unlock_func"
	case WaitGroup:
		return "wait_group"
	case Queue:
		return "fifo_queue"
	case LifoQueue:
		return "lifo_queue"
	default:
		return ""
	}
}

type ActionFlag int

const (
	VariableDefinition ActionFlag = 0x1
	VariableAssignment ActionFlag = 0x2
	VariableAdd        ActionFlag = 0x4
	VariableSubtract   ActionFlag = 0x8
	VariableMultiply   ActionFlag = 0x10
	VariableDivide     ActionFlag = 0x20
	VariableUsage      ActionFlag = 0x40
	MutexLock          ActionFlag = 0x80
	MutexUnlock        ActionFlag = 0x100
	Wait               ActionFlag = 0x200
	QueueFull          ActionFlag = 0x400
	QueueEmpty         ActionFlag = 0x800
	QueueSize          ActionFlag = 0x1000
	QueuePut           ActionFlag = 0x2000
	QueueGet           ActionFlag = 0x4000
)

type Variable struct {
	ID           string
	Type         VariableType
	SubType      VariableType
	Name         string
	Tag          string
	DefaultValue any
}

func (v *Variable) String() string {
	return v.Name
}

type Statement struct {
	Line           int
	Indentation    string
	Actions        ActionFlag
	TargetVariable *Variable
	UsedVariables  []*Variable
	OriginalCode   string
	ModifiedCode   string
	OriginalLValue string
	OriginalRValue string
	ModifiedRValue string
	ShouldSkip     bool
}

type Module struct {
	Name             string
	AbsolutePath     string
	RelativePath     string
	PackageDirectory string
	RequireParens    bool
	VariablePrefix   string
	VariableSuffix   string
	Statements       []*Statement
}

func (m *Module) GetVariables() []*Variable {
	vars := make([]*Variable, 0)
	for _, statement := range m.Statements {
		if statement.Actions == VariableDefinition && statement.TargetVariable != nil {
			vars = append(vars, statement.TargetVariable)
		}
	}
	return vars
}

func (m *Module) GetVariableByName(name string) *Variable {
	for _, statement := range m.Statements {
		if statement.Actions == VariableDefinition && statement.TargetVariable != nil {
			if statement.TargetVariable.Name == name {
				return statement.TargetVariable
			}
		}
	}

	return nil
}

func (m *Module) GetVariableByID(id string) *Variable {
	for _, statement := range m.Statements {
		if statement.Actions == VariableDefinition && statement.TargetVariable != nil {
			if statement.TargetVariable.ID == id {
				return statement.TargetVariable
			}
		}
	}

	return nil
}

func (m *Module) GetStatement(line int) *Statement {
	for _, statement := range m.Statements {
		if statement.Line == line {
			return statement
		}
	}

	return nil
}

type Package []*Module

func (p Package) Directory() string {
	if len(p) == 0 {
		return ""
	}

	return p[0].PackageDirectory
}

func (p Package) GetVariables() []*Variable {
	vars := make([]*Variable, 0)
	for _, mod := range p {
		vars = append(vars, mod.GetVariables()...)
	}
	return vars
}

func (p Package) GetVariableByID(id string) *Variable {
	for _, mod := range p {
		variable := mod.GetVariableByID(id)
		if variable != nil {
			return variable
		}
	}
	return nil
}

func (a ActionFlag) Contains(action ActionFlag) bool {
	return a&action == action
}
