package receptor

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry/gunk/urljoiner"
	"github.com/tedsuo/rata"
)

type Client interface {
	CreateTask(TaskCreateRequest) error
	GetAllTasks() ([]TaskResponse, error)
	GetAllTasksByDomain(domain string) ([]TaskResponse, error)
	GetTask(taskId string) (TaskResponse, error)
	DeleteTask(taskId string) error

	CreateDesiredLRP(DesiredLRPCreateRequest) error
	UpdateDesiredLRP(processGuid string, update DesiredLRPUpdateRequest) error
	DeleteDesiredLRP(processGuid string) error
	GetAllDesiredLRPs() ([]DesiredLRPResponse, error)
	GetAllDesiredLRPsByDomain(domain string) ([]DesiredLRPResponse, error)
	GetDesiredLRPByProcessGuid(processGuid string) (DesiredLRPResponse, error)
}

func NewClient(addr, user, password string) Client {
	return &client{
		user:       user,
		password:   password,
		httpClient: &http.Client{},
		reqGen:     rata.NewRequestGenerator(urljoiner.Join("http://", addr), Routes),
	}
}

type client struct {
	user       string
	password   string
	httpClient *http.Client
	reqGen     *rata.RequestGenerator
}

func (c *client) CreateTask(request TaskCreateRequest) error {
	return c.doRequest(CreateTaskRoute, nil, request, nil)
}

func (c *client) GetAllTasks() ([]TaskResponse, error) {
	tasks := []TaskResponse{}
	err := c.doRequest(GetAllTasksRoute, nil, nil, &tasks)
	return tasks, err
}

func (c *client) GetAllTasksByDomain(domain string) ([]TaskResponse, error) {
	tasks := []TaskResponse{}
	err := c.doRequest(GetAllTasksByDomainRoute, rata.Params{"domain": domain}, nil, &tasks)
	return tasks, err
}

func (c *client) GetTask(taskId string) (TaskResponse, error) {
	task := TaskResponse{}
	err := c.doRequest(GetTaskRoute, rata.Params{"task_guid": taskId}, nil, &task)
	return task, err
}

func (c *client) DeleteTask(taskId string) error {
	return c.doRequest(DeleteTaskRoute, rata.Params{"task_guid": taskId}, nil, nil)
}

func (c *client) CreateDesiredLRP(req DesiredLRPCreateRequest) error {
	return c.doRequest(CreateDesiredLRPRoute, nil, req, nil)
}

func (c *client) UpdateDesiredLRP(processGuid string, req DesiredLRPUpdateRequest) error {
	return c.doRequest(UpdateDesiredLRPRoute, rata.Params{"process_guid": processGuid}, req, nil)
}

func (c *client) DeleteDesiredLRP(processGuid string) error {
  return c.doRequest(DeleteDesiredLRPRoute, rata.Params{"process_guid": processGuid}, nil, nil)
}

func (c *client) GetAllDesiredLRPs() ([]DesiredLRPResponse, error) {
	var desiredLRPs []DesiredLRPResponse
	err := c.doRequest(GetAllDesiredLRPsRoute, nil, nil, &desiredLRPs)
	return desiredLRPs, err
}

func (c *client) GetAllDesiredLRPsByDomain(domain string) ([]DesiredLRPResponse, error) {
	var desiredLRPs []DesiredLRPResponse
	err := c.doRequest(GetAllDesiredLRPsByDomainRoute, rata.Params{"domain": domain}, nil, &desiredLRPs)
	return desiredLRPs, err
}

func (c *client) GetDesiredLRPByProcessGuid(processGuid string) (DesiredLRPResponse, error) {
	var desiredLRP DesiredLRPResponse
	err := c.doRequest(GetDesiredLRPByProcessGuidRoute, rata.Params{"process_guid": processGuid}, nil, &desiredLRP)
	return desiredLRP, err
}

func (c *client) doRequest(requestName string, params rata.Params, request, response interface{}) error {
	requestJson, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := c.reqGen.CreateRequest(requestName, params, bytes.NewReader(requestJson))
	if err != nil {
		return err
	}

	req.ContentLength = int64(len(requestJson))
	req.SetBasicAuth(c.user, c.password)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode > 299 {
		errResponse := Error{}
		json.NewDecoder(res.Body).Decode(&errResponse)
		return errResponse
	}

	if response != nil {
		return json.NewDecoder(res.Body).Decode(&response)
	} else {
		return nil
	}
}
