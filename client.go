package receptor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/tedsuo/rata"
)

type Client interface {
	CreateTask(TaskCreateRequest) error
	Tasks() ([]TaskResponse, error)
	TasksByDomain(domain string) ([]TaskResponse, error)
	GetTask(taskId string) (TaskResponse, error)
	DeleteTask(taskId string) error
	CancelTask(taskId string) error

	CreateDesiredLRP(DesiredLRPCreateRequest) error
	GetDesiredLRP(processGuid string) (DesiredLRPResponse, error)
	UpdateDesiredLRP(processGuid string, update DesiredLRPUpdateRequest) error
	DeleteDesiredLRP(processGuid string) error
	DesiredLRPs() ([]DesiredLRPResponse, error)
	DesiredLRPsByDomain(domain string) ([]DesiredLRPResponse, error)

	ActualLRPs() ([]ActualLRPResponse, error)
	ActualLRPsByDomain(domain string) ([]ActualLRPResponse, error)
	ActualLRPsByProcessGuid(processGuid string) ([]ActualLRPResponse, error)
	ActualLRPByProcessGuidAndIndex(processGuid string, index int) (ActualLRPResponse, error)
	KillActualLRPByProcessGuidAndIndex(processGuid string, index int) error

	Cells() ([]CellResponse, error)

	BumpFreshDomain(FreshDomainBumpRequest) error
	FreshDomains() ([]FreshDomainResponse, error)
}

func NewClient(url string) Client {
	return &client{
		httpClient: &http.Client{},
		reqGen:     rata.NewRequestGenerator(url, Routes),
	}
}

type client struct {
	httpClient *http.Client
	reqGen     *rata.RequestGenerator
}

func (c *client) CreateTask(request TaskCreateRequest) error {
	return c.doRequest(CreateTaskRoute, nil, nil, request, nil)
}

func (c *client) Tasks() ([]TaskResponse, error) {
	tasks := []TaskResponse{}
	err := c.doRequest(TasksRoute, nil, nil, nil, &tasks)
	return tasks, err
}

func (c *client) TasksByDomain(domain string) ([]TaskResponse, error) {
	tasks := []TaskResponse{}
	err := c.doRequest(TasksByDomainRoute, rata.Params{"domain": domain}, nil, nil, &tasks)
	return tasks, err
}

func (c *client) GetTask(taskId string) (TaskResponse, error) {
	task := TaskResponse{}
	err := c.doRequest(GetTaskRoute, rata.Params{"task_guid": taskId}, nil, nil, &task)
	return task, err
}

func (c *client) DeleteTask(taskId string) error {
	return c.doRequest(DeleteTaskRoute, rata.Params{"task_guid": taskId}, nil, nil, nil)
}

func (c *client) CancelTask(taskId string) error {
	return c.doRequest(CancelTaskRoute, rata.Params{"task_guid": taskId}, nil, nil, nil)
}

func (c *client) CreateDesiredLRP(req DesiredLRPCreateRequest) error {
	return c.doRequest(CreateDesiredLRPRoute, nil, nil, req, nil)
}

func (c *client) GetDesiredLRP(processGuid string) (DesiredLRPResponse, error) {
	var desiredLRP DesiredLRPResponse
	err := c.doRequest(GetDesiredLRPRoute, rata.Params{"process_guid": processGuid}, nil, nil, &desiredLRP)
	return desiredLRP, err
}

func (c *client) UpdateDesiredLRP(processGuid string, req DesiredLRPUpdateRequest) error {
	return c.doRequest(UpdateDesiredLRPRoute, rata.Params{"process_guid": processGuid}, nil, req, nil)
}

func (c *client) DeleteDesiredLRP(processGuid string) error {
	return c.doRequest(DeleteDesiredLRPRoute, rata.Params{"process_guid": processGuid}, nil, nil, nil)
}

func (c *client) DesiredLRPs() ([]DesiredLRPResponse, error) {
	var desiredLRPs []DesiredLRPResponse
	err := c.doRequest(DesiredLRPsRoute, nil, nil, nil, &desiredLRPs)
	return desiredLRPs, err
}

func (c *client) DesiredLRPsByDomain(domain string) ([]DesiredLRPResponse, error) {
	var desiredLRPs []DesiredLRPResponse
	err := c.doRequest(DesiredLRPsByDomainRoute, rata.Params{"domain": domain}, nil, nil, &desiredLRPs)
	return desiredLRPs, err
}

func (c *client) ActualLRPs() ([]ActualLRPResponse, error) {
	var actualLRPs []ActualLRPResponse
	err := c.doRequest(ActualLRPsRoute, nil, nil, nil, &actualLRPs)
	return actualLRPs, err
}

func (c *client) ActualLRPsByDomain(domain string) ([]ActualLRPResponse, error) {
	var actualLRPs []ActualLRPResponse
	err := c.doRequest(ActualLRPsByDomainRoute, rata.Params{"domain": domain}, nil, nil, &actualLRPs)
	return actualLRPs, err
}

func (c *client) ActualLRPsByProcessGuid(processGuid string) ([]ActualLRPResponse, error) {
	var actualLRPs []ActualLRPResponse
	err := c.doRequest(ActualLRPsByProcessGuidRoute, rata.Params{"process_guid": processGuid}, nil, nil, &actualLRPs)
	return actualLRPs, err
}

func (c *client) ActualLRPByProcessGuidAndIndex(processGuid string, index int) (ActualLRPResponse, error) {
	var actualLRP ActualLRPResponse
	err := c.doRequest(ActualLRPByProcessGuidAndIndexRoute, rata.Params{"process_guid": processGuid, "index": strconv.Itoa(index)}, nil, nil, &actualLRP)
	return actualLRP, err
}

func (c *client) KillActualLRPByProcessGuidAndIndex(processGuid string, index int) error {
	err := c.doRequest(KillActualLRPByProcessGuidAndIndexRoute, rata.Params{"process_guid": processGuid, "index": strconv.Itoa(index)}, nil, nil, nil)
	return err
}

func (c *client) Cells() ([]CellResponse, error) {
	var cells []CellResponse
	err := c.doRequest(CellsRoute, nil, nil, nil, &cells)
	return cells, err
}

func (c *client) BumpFreshDomain(req FreshDomainBumpRequest) error {
	return c.doRequest(BumpFreshDomainRoute, nil, nil, req, nil)
}

func (c *client) FreshDomains() ([]FreshDomainResponse, error) {
	var freshDomains []FreshDomainResponse
	err := c.doRequest(FreshDomainsRoute, nil, nil, nil, &freshDomains)
	return freshDomains, err
}

func (c *client) doRequest(requestName string, params rata.Params, queryParams url.Values, request, response interface{}) error {
	requestJson, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := c.reqGen.CreateRequest(requestName, params, bytes.NewReader(requestJson))
	if err != nil {
		return err
	}

	req.URL.RawQuery = queryParams.Encode()
	req.ContentLength = int64(len(requestJson))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		errResponse := Error{}
		json.NewDecoder(res.Body).Decode(&errResponse)
		return errResponse
	}

	if response != nil {
		return json.NewDecoder(res.Body).Decode(response)
	} else {
		return nil
	}
}
