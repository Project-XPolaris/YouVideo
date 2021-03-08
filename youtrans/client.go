package youtrans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/projectxpolaris/youvideo/config"
	"net/http"
)

var DefaultYouTransClient = YouTransClient{}

type YouTransClient struct {
}
type CreateTaskRequestBody struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Format string `json:"format"`
	Codec  string `json:"codec"`
}

type TaskResponse struct {
	Id      string  `json:"id"`
	Process float64 `json:"process"`
	Input   string  `json:"input"`
	Output  string  `json:"output"`
	Status  string  `json:"status"`
}

func (c *YouTransClient) GetUrl(path string) string {
	return fmt.Sprintf("%s/%s", config.Instance.YoutransURL, path)
}
func (c *YouTransClient) CreateNewTask(body *CreateTaskRequestBody) (*TaskResponse, error) {
	var responseBody TaskResponse
	err := c.makePOSTRequest("tasks", body, &responseBody)
	if err != nil {
		return nil, err
	}
	return &responseBody, err
}

func (c *YouTransClient) makePOSTRequest(url string, data interface{}, responseBody interface{}) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	client := http.Client{}
	request, err := http.NewRequest("POST", c.GetUrl(url), bytes.NewBuffer(rawData))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	return err
}

func (c *YouTransClient) makeGETRequest(url string, responseBody interface{}) error {
	response, err := http.Get(c.GetUrl(url))
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	return err
}

type TaskListResponse struct {
	List []TaskResponse `json:"list"`
}

func (c *YouTransClient) GetTaskList() (*TaskListResponse, error) {
	var responseBody TaskListResponse
	err := c.makeGETRequest("/tasks", &responseBody)
	if err != nil {
		return nil, err
	}
	return &responseBody, err
}
