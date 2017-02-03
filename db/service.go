package db

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

type serviceKeyResponse struct {
	Resources []serviceKeyResource `json:"resources"`
}

type serviceKeyResource struct {
	Entity struct {
		Credentials credentialsJSON `json:"credentials"`
	} `json:"entity"`
}

type Credentials interface {
	GetDBName() string
	GetHost() string
	GetUsername() string
	GetPassword() string
	GetPort() string
}

type credentialsJSON struct {
	// these groups of fields should be interchangeable
	DBName string `json:"db_name"`
	Name   string `json:"name"`

	Host     string `json:"host"`
	Hostname string `json:"hostname"`
	HostName string `json:"host_name"`

	Username string `json:"username"`
	UserName string `json:"user_name"`
	User     string `json:"user"`

	Password string `json:"password"`
	Pass     string `json:"pass"`

	Port int `json:"port"`
}

func (c credentialsJSON) GetDBName() string {
	if c.Name != "" {
		return c.Name
	}
	return c.DBName
}

func (c credentialsJSON) GetHost() string {
	if c.Host != "" {
		return c.Host
	}
	if c.HostName != "" {
		return c.HostName
	}
	return c.Hostname
}

func (c credentialsJSON) GetUsername() string {
	if c.Username != "" {
		return c.Username
	}
	if c.UserName != "" {
		return c.UserName
	}
	return c.User
}

func (c credentialsJSON) GetPassword() string {
	if c.Pass != "" {
		return c.Pass
	}
	return c.Password
}

func (c credentialsJSON) GetPort() string {
	return strconv.Itoa(c.Port)
}

func CredentialsFromJSON(body string) (creds Credentials, err error) {
	serviceKeyResponse := serviceKeyResponse{}
	err = json.Unmarshal([]byte(body), &serviceKeyResponse)
	if err != nil {
		return
	}
	creds = serviceKeyResponse.Resources[0].Entity.Credentials

	return
}

type Services map[string][]Service

type Service struct {
	Name        string
	Plan        string
	Label       string
	Credentials credentialsJSON
}

func LoadServiceCredentials(serviceName string) (creds Credentials, err error) {
	if os.Getenv("VCAP_SERVICES") == "" {
		return credentialsJSON{}, errors.New("no p-mysql services is not added")
	}
	var services Services
	err = json.Unmarshal([]byte(os.Getenv("VCAP_SERVICES")), &services)

	if err != nil {
		return credentialsJSON{}, err
	}
	value := services[serviceName][0]

	return value.Credentials, err
}
