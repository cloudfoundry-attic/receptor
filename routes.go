package receptor

import "github.com/tedsuo/rata"

const (
	CreateTask          = "CreateTask"
	GetAllTasks         = "GetAllTasks"
	GetAllTasksByDomain = "GetAllTasksByDomain"
	GetTask             = "GetTask"
	DeleteTask          = "DeleteTask"

	CreateDesiredLRP = "CreateDesiredLRP"
)

var Routes = rata.Routes{
	{Path: "/tasks", Method: "POST", Name: CreateTask},
	{Path: "/tasks", Method: "GET", Name: GetAllTasks},
	{Path: "/domains/:domain/tasks", Method: "GET", Name: GetAllTasksByDomain},
	{Path: "/tasks/:task_guid", Method: "GET", Name: GetTask},
	{Path: "/tasks/:task_guid", Method: "DELETE", Name: DeleteTask},

	{Path: "/desired_lrps", Method: "POST", Name: CreateDesiredLRP},
}
