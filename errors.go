package main

import "errors"

var (
	ErrNoCommandFoundCommands       = errors.New("no command found in commands")
	ErrArgProvidedButNotUsed        = errors.New("argument provided but not used in the template")
	ErrArgUsedInTemplateNotProvided = errors.New("argument used in template but not provided in args")
	ErrNoCommandFoundHeading        = errors.New("no command found in heading")
	ErrNoLauncherDefined            = errors.New("no launcher defined for infostring")
	ErrNoInfostringOrShebang        = errors.New("no infostring and no shebang defined")
	ErrDuplicateCommand             = errors.New("duplicate command found")
)
