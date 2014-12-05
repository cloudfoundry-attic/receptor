package receptor

import "github.com/tedsuo/rata"

const (
	// Tasks
	CreateTaskRoute    = "CreateTask"
	TasksRoute         = "Tasks"
	TasksByDomainRoute = "TasksByDomain"
	GetTaskRoute       = "GetTask"
	DeleteTaskRoute    = "DeleteTask"
	CancelTaskRoute    = "CancelTask"

	// DesiredLRPs
	CreateDesiredLRPRoute    = "CreateDesiredLRP"
	GetDesiredLRPRoute       = "GetDesiredLRP"
	UpdateDesiredLRPRoute    = "UpdateDesiredLRP"
	DeleteDesiredLRPRoute    = "DeleteDesiredLRP"
	DesiredLRPsRoute         = "DesiredLRPs"
	DesiredLRPsByDomainRoute = "DesiredLRPsByDomain"

	// ActualLRPs
	ActualLRPsRoute                         = "ActualLRPs"
	ActualLRPsByDomainRoute                 = "ActualLRPsByDomain"
	ActualLRPsByProcessGuidRoute            = "ActualLRPsByProcessGuid"
	ActualLRPByProcessGuidAndIndexRoute     = "ActualLRPByProcessGuidAndIndex"
	KillActualLRPByProcessGuidAndIndexRoute = "KillActualLRPByProcessGuidAndIndex"

	// Cells
	CellsRoute = "Cells"

	// Fresh domains
	BumpFreshDomainRoute = "BumpFreshDomain"
	FreshDomainsRoute    = "FreshDomains"
)

var Routes = rata.Routes{
	// Tasks
	{Path: "/v1/tasks", Method: "POST", Name: CreateTaskRoute},
	{Path: "/v1/tasks", Method: "GET", Name: TasksRoute},
	{Path: "/v1/domains/:domain/tasks", Method: "GET", Name: TasksByDomainRoute},
	{Path: "/v1/tasks/:task_guid", Method: "GET", Name: GetTaskRoute},
	{Path: "/v1/tasks/:task_guid", Method: "DELETE", Name: DeleteTaskRoute},
	{Path: "/v1/tasks/:task_guid/cancel", Method: "POST", Name: CancelTaskRoute},

	// DesiredLRPs
	{Path: "/v1/desired_lrps", Method: "POST", Name: CreateDesiredLRPRoute},
	{Path: "/v1/desired_lrps/:process_guid", Method: "GET", Name: GetDesiredLRPRoute},
	{Path: "/v1/desired_lrps/:process_guid", Method: "PUT", Name: UpdateDesiredLRPRoute},
	{Path: "/v1/desired_lrps/:process_guid", Method: "DELETE", Name: DeleteDesiredLRPRoute},
	{Path: "/v1/desired_lrps", Method: "GET", Name: DesiredLRPsRoute},
	{Path: "/v1/domains/:domain/desired_lrps", Method: "GET", Name: DesiredLRPsByDomainRoute},

	// ActualLRPs
	{Path: "/v1/actual_lrps", Method: "GET", Name: ActualLRPsRoute},
	{Path: "/v1/domains/:domain/actual_lrps", Method: "GET", Name: ActualLRPsByDomainRoute},
	{Path: "/v1/actual_lrps/:process_guid", Method: "GET", Name: ActualLRPsByProcessGuidRoute},
	{Path: "/v1/actual_lrps/:process_guid/index/:index", Method: "GET", Name: ActualLRPByProcessGuidAndIndexRoute},
	{Path: "/v1/actual_lrps/:process_guid/index/:index", Method: "DELETE", Name: KillActualLRPByProcessGuidAndIndexRoute},

	// Cells
	{Path: "/v1/cells", Method: "GET", Name: CellsRoute},

	// Fresh domains
	{Path: "/v1/fresh_domains", Method: "POST", Name: BumpFreshDomainRoute},
	{Path: "/v1/fresh_domains", Method: "GET", Name: FreshDomainsRoute},
}
