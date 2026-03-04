package migrations

type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}
