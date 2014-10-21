package receptor

import "github.com/tedsuo/rata"

const (
	CreateTask  = "CreateTask"
	GetAllTasks = "GetAllTasks"
)

var Routes = rata.Routes{
	{Path: "/tasks", Method: "POST", Name: CreateTask},
	{Path: "/tasks", Method: "GET", Name: GetAllTasks},
}
