package types

type Task struct {
	// step
	Init   Step
	Run    Code
	Verify Validator

	Runtime string
}

type Code struct {
}

type Validator struct {
	Custom  *Step
	Default *DefaultValidator
}

type DefaultValidator struct {
	Name string
}

type Report struct {
	Status   string
	Result   string
	Messages []string
}
