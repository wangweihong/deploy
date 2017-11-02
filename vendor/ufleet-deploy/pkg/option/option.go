package option

type CreateOption struct {
	Comment string
	Name    *string //如果为空,取k8s resource名
	User    string
}

type DeleteOption struct {
	Group     string
	Workspace string
	Name      string
}
