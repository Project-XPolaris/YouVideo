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
	return fmt.Sprintf("%s/%s", config.AppConfig.YoutransURL, path)
}
func (c *YouTransClient) CreateNewTask(body *CreateTaskRequestBody) (*TaskResponse, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	response, err := http.Post(c.GetUrl("tasks"), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	var responseBody TaskResponse
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	return &responseBody, err
}
