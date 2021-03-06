package uaa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pankrator/payment/auth"
	"github.com/pankrator/payment/web"
)

type UAAConfig struct {
	Auth *auth.Config
	URL  string
}

type UAAClient struct {
	client auth.Client
	config *UAAConfig
}

func NewClient(config *UAAConfig) (*UAAClient, error) {
	info, err := auth.GetInfo(config.URL)
	if err != nil {
		return nil, err
	}
	config.Auth.TokenEndpoint = info.TokenEndpoint
	client := auth.New(config.Auth)

	return &UAAClient{
		client: client,
		config: config,
	}, nil
}

type emailValue struct {
	Value string `json:"value"`
}

type createUserRequest struct {
	Emails   []emailValue `json:"emails"`
	UserName string       `json:"userName"`
	Verified bool         `json:"verified"`
	Active   bool         `json:"active"`
	Password string       `json:"password"`
}

type Group struct {
	DisplayName string `json:"displayName"`
	ID          string `json:"id"`
}

func (uc *UAAClient) CreateUser(ctx context.Context, username, email, password string) (string, error) {
	createUserBytes, err := json.Marshal(&createUserRequest{
		Emails: []emailValue{
			{
				Value: email,
			},
		},
		UserName: username,
		Password: password,
		Active:   true,
		Verified: true,
	})
	if err != nil {
		return "", fmt.Errorf("could not marshal create user request: %s", err)
	}
	reader := bytes.NewReader(createUserBytes)

	log.Printf("Requesting UAA to endpoint %s", uc.config.URL+"/Users")
	req, err := http.NewRequest(http.MethodPost, uc.config.URL+"/Users", reader)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return "", err
	}

	resp, err := uc.client.Do(req)
	if err != nil {
		return "", err
	}
	var result map[string]interface{}
	err = web.BodyToObject(resp.Body, &result)
	if err != nil {
		return "", err
	}

	switch resp.StatusCode {
	case http.StatusConflict:
		log.Printf("User with name %s already exists, skipping it", username)
		return "", nil
	case http.StatusCreated:
		return result["id"].(string), nil
	default:
		return "", fmt.Errorf("uaa responded with status code: %d and error: %s", resp.StatusCode, result["error_description"].(string))
	}
}

func (uc *UAAClient) GetGroup(ctx context.Context, displayName string) (*Group, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/Groups", uc.config.URL), nil)
	query := req.URL.Query()
	query.Add("filter", fmt.Sprintf(`displayName eq "%s"`, displayName))
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = query.Encode()

	resp, err := uc.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("uaa responded with status code: %d", resp.StatusCode)
	}

	groups := &struct {
		Resources []*Group `json:"resources"`
	}{}
	if err = web.BodyToObject(resp.Body, groups); err != nil {
		return nil, err
	}
	log.Printf("Found group with name: %s", groups.Resources[0].DisplayName)
	return groups.Resources[0], nil
}

func (uc *UAAClient) AddUserToGroup(ctx context.Context, userID string, group *Group) (bool, error) {
	log.Printf("Adding user %s to group %s", userID, group.DisplayName)
	body := map[string]interface{}{}
	body["members"] = []map[string]interface{}{
		{
			"value":  userID,
			"type":   "USER",
			"origin": "uaa",
		},
	}
	body["displayName"] = group.DisplayName

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return false, err
	}

	reader := bytes.NewReader(bodyBytes)
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/Groups/%s", uc.config.URL, group.ID), reader)
	if err != nil {
		return false, err
	}
	req.Header.Set("If-Match", "*")
	req.Header.Set("Content-Type", "application/json")

	resp, err := uc.client.Do(req)
	if err != nil {
		return false, err
	}

	var response map[string]interface{}
	if err = web.BodyToObject(resp.Body, &response); err != nil {
		return false, err
	}
	switch resp.StatusCode {
	case http.StatusConflict:
		log.Printf("User is already a member of the group")
		return false, nil
	case http.StatusOK:
		return true, nil
	default:
		return false, fmt.Errorf("uaa responded with status code: %d and body %+v", resp.StatusCode, response)
	}
}
