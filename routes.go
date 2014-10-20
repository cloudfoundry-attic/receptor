package receptor

import "github.com/tedsuo/rata"

const (
	CreateTask = "CreateTask"
)

var Routes = rata.Routes{
	{Path: "/tasks", Method: "POST", Name: CreateTask},
}
