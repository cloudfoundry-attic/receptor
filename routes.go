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
	{Path: "/v1/tasks", Method: "POST", Name: CreateTaskRoute},
	{Path: "/v1/tasks", Method: "GET", Name: GetAllTasksRoute},
	{Path: "/v1/domains/:domain/tasks", Method: "GET", Name: GetAllTasksByDomainRoute},
	{Path: "/v1/tasks/:task_guid", Method: "GET", Name: GetTaskRoute},
	{Path: "/v1/tasks/:task_guid", Method: "DELETE", Name: DeleteTaskRoute},

	// DesiredLRPs
	{Path: "/v1/desired_lrps", Method: "POST", Name: CreateDesiredLRPRoute},
	{Path: "/v1/desired_lrps/:process_guid", Method: "GET", Name: GetDesiredLRPRoute},
	{Path: "/v1/desired_lrps/:process_guid", Method: "PUT", Name: UpdateDesiredLRPRoute},
	{Path: "/v1/desired_lrps/:process_guid", Method: "DELETE", Name: DeleteDesiredLRPRoute},
	{Path: "/v1/desired_lrps", Method: "GET", Name: GetAllDesiredLRPsRoute},
	{Path: "/v1/domains/:domain/desired_lrps", Method: "GET", Name: GetAllDesiredLRPsByDomainRoute},

	// ActualLRPs
	{Path: "/v1/actual_lrps", Method: "GET", Name: GetAllActualLRPsRoute},
	{Path: "/v1/domains/:domain/actual_lrps", Method: "GET", Name: GetAllActualLRPsByDomainRoute},
	{Path: "/v1/desired_lrps/:process_guid/actual_lrps", Method: "GET", Name: GetAllActualLRPsByProcessGuidRoute},
	{Path: "/v1/desired_lrps/:process_guid/actual_lrps", Method: "DELETE", Name: KillActualLRPsByProcessGuidAndIndexRoute},

	// Cells
	{Path: "/v1/cells", Method: "GET", Name: CellsRoute},

	// Fresh domains
	{Path: "/v1/fresh_domains", Method: "POST", Name: CreateFreshDomainRoute},
	{Path: "/v1/fresh_domains", Method: "GET", Name: FreshDomainsRoute},
}
