package args

func (e *UserCreate) Help() string {
	return `
create
`
}

func (e *User) Help() string {
	return (`
Create and manage users.

create                Create a new user interactively
    --allFields       Prompt for all possible user fields instead of setting some default values
	`)
}
