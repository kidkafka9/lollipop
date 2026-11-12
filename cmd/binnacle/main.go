package main

const (
	applicationName  = "binnacle"
	notProvidedValue = "[not-provided]"
)

type model struct {
	width, height int
}

type Model interface {
	Init() error

	Update() error

	View() error
}
