package request

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Send[T any](url, method string, body any) (*T, error) {
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := (&http.Client{}).Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf("Сервер ответил кодом %d", response.StatusCode)
		return nil, err
	}

	var data T
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
