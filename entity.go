package rubik

type punchInEn struct {
	Entity
	Workspace string `rubik:"body"`
	Service   string `rubik:"body"`
	Host      string `rubik:"body"`
	Port      string `rubik:"body"`
}
