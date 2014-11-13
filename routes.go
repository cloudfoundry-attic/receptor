package receptor

import "github.com/tedsuo/rata"

const (
	// Tasks
	CreateTaskRoute          = "CreateTask"
	GetAllTasksRoute         = "GetAllTasks"
	GetAllTasksByDomainRoute = "GetAllTasksByDomain"
	GetTaskRoute             = "GetTask"
	DeleteTaskRoute          = "DeleteTask"

	// DesiredLRPs
	CreateDesiredLRPRoute          = "CreateDesiredLRP"
	GetDesiredLRPRoute             = "GetDesiredLRP"
	UpdateDesiredLRPRoute          = "UpdateDesiredLRP"
	DeleteDesiredLRPRoute          = "DeleteDesiredLRP"
	GetAllDesiredLRPsRoute         = "GetAllDesiredLRPs"
	GetAllDesiredLRPsByDomainRoute = "GetAllDesiredLRPsByDomain"

	// ActualLRPs
	GetAllActualLRPsRoute                    = "GetAllActualLRPs"
	GetAllActualLRPsByDomainRoute            = "GetAllActualLRPsByDomain"
	GetAllActualLRPsByProcessGuidRoute       = "GetAllActualLRPsByProcessGuid"
	KillActualLRPsByProcessGuidAndIndexRoute = "KillActualLRPsByProcessGuidAndIndex"

	// Cells
	CellsRoute = "Cells"

	// Fresh domains
	CreateFreshDomainRoute = "CreateFreshDomain"
	FreshDomainsRoute      = "FreshDomains"
)

var Routes = rata.Routes{
	// Tasks
	{Path: "/tasks", Method: "POST", Name: CreateTaskRoute},
	{Path: "/tasks", Method: "GET", Name: GetAllTasksRoute},
	{Path: "/domains/:domain/tasks", Method: "GET", Name: GetAllTasksByDomainRoute},
	{Path: "/tasks/:task_guid", Method: "GET", Name: GetTaskRoute},
	{Path: "/tasks/:task_guid", Method: "DELETE", Name: DeleteTaskRoute},

	// DesiredLRPs
	{Path: "/desired_lrps", Method: "POST", Name: CreateDesiredLRPRoute},
	{Path: "/desired_lrps/:process_guid", Method: "GET", Name: GetDesiredLRPRoute},
	{Path: "/desired_lrps/:process_guid", Method: "PUT", Name: UpdateDesiredLRPRoute},
	{Path: "/desired_lrps/:process_guid", Method: "DELETE", Name: DeleteDesiredLRPRoute},
	{Path: "/desired_lrps", Method: "GET", Name: GetAllDesiredLRPsRoute},
	{Path: "/domains/:domain/desired_lrps", Method: "GET", Name: GetAllDesiredLRPsByDomainRoute},

	// ActualLRPs
	{Path: "/actual_lrps", Method: "GET", Name: GetAllActualLRPsRoute},
	{Path: "/domains/:domain/actual_lrps", Method: "GET", Name: GetAllActualLRPsByDomainRoute},
	{Path: "/desired_lrps/:process_guid/actual_lrps", Method: "GET", Name: GetAllActualLRPsByProcessGuidRoute},
	{Path: "/desired_lrps/:process_guid/actual_lrps", Method: "DELETE", Name: KillActualLRPsByProcessGuidAndIndexRoute},

	// Cells
	{Path: "/cells", Method: "GET", Name: CellsRoute},

	// Fresh domains
	{Path: "/fresh_domains", Method: "POST", Name: CreateFreshDomainRoute},
	{Path: "/fresh_domains", Method: "GET", Name: FreshDomainsRoute},
}
