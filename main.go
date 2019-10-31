package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/christianwoehrle/prometheus-testclient/prometheus_cw"

	"github.com/christianwoehrle/prometheus-testclient/grafana_dtos_cw"
	yaml2 "gopkg.in/yaml.v2"
)

type Datasource struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Id    int64  `json:"id"`
	Tests []Test `yaml:"tests"`
}

type Datasources struct {
	Datasource []Datasource `yaml:"datasources"`
}

type Test struct {
	ProxyQuery string `yaml:"proxyQuery"`
}

const LOKI_QUERYPATH = "/api/prom/query"
const PROM_QUERYPATH = "/api/v1/query"

func Query(user string, pass string, addr string, path string, rawQuery string) ([]byte, error) {
	scheme := "http"
	u := url.URL{
		Scheme:   scheme,
		User:     url.UserPassword(user, pass),
		Host:     addr,
		Path:     path,
		RawQuery: rawQuery,
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(user, pass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("error closing body", err)
		}
	}()

	if resp.StatusCode/100 != 2 {
		buf, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from server: %s (%v)", string(buf), err)
	}

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		//bodyString := string(bodyBytes)
		//fmt.Println(bodyString)
		return bodyBytes, nil

	}
	return nil, fmt.Errorf("error response from server: %d ", resp.StatusCode)

}

func main() {

	filename, _ := filepath.Abs("./check.yaml")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("Failed to read config-file")
		panic(err)
	}

	var datasourceSpecs Datasources
	err = yaml2.Unmarshal(yamlFile, &datasourceSpecs)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	user := "admin"
	passwd := "admin"
	grafanaAdress := "localhost:8080"

	res, err := Query(user, passwd, grafanaAdress, "/api/datasources", "")

	var datasourcelist grafana_dtos_cw.DataSourceList
	err = json.Unmarshal(res, &datasourcelist)

	for i, _ := range datasourceSpecs.Datasource {
		res, err = Query(user, passwd, grafanaAdress, "/api/datasources/id/"+datasourceSpecs.Datasource[i].Name, "")
		if err != nil {
			fmt.Println("Fail: datasource " + datasourceSpecs.Datasource[i].Name + " not found")
		}
		var datasourceId grafana_dtos_cw.DataSourceID
		err = json.Unmarshal(res, &datasourceId)
		if err != nil {
			fmt.Println("Fail: datasourceid " + datasourceSpecs.Datasource[i].Name + " not found")
		}
		datasourceSpecs.Datasource[i].Id = datasourceId.Id
		fmt.Println("Found datasource in Grafana:" + datasourceSpecs.Datasource[i].Name)
	}

	res, err = Query(user, passwd, grafanaAdress, "/api/health", "")
	var health map[string]interface{}
	err = json.Unmarshal(res, &health)
	fmt.Println("Check /api/health, database:" + health["database"].(string))

	/*
		{
			"commit": "67bad72",
			"database": "ok",
			"version": "6.3.5"
		}
	*/

	for _, datasource := range datasourceSpecs.Datasource {
		for _, test := range datasource.Tests {
			if datasource.Id == 0 {
				fmt.Println("No ID for Datasource <" + datasource.Name + ">, Skip Queries")
			} else {

				if datasource.Type == "Prometheus" {
					path := "/api/datasources/proxy/" + strconv.Itoa(int(datasource.Id)) + PROM_QUERYPATH

					res, err = Query(user, passwd, grafanaAdress, path, test.ProxyQuery)
					var promApiResponse prometheus_cw.ApiResponse
					err = json.Unmarshal(res, &promApiResponse)

					//fmt.Println(err, string(res), promApiResponse)
					if err != nil {
						fmt.Println("Query failed: " + datasource.Name + " --> " + test.ProxyQuery + " --> " + err.Error())
					} else {

						if promApiResponse.Status == "success" {
							fmt.Println("Query ok: " + datasource.Name + " --> " + test.ProxyQuery)
						} else {
							fmt.Println("Query failed: " + datasource.Name + " --> " + test.ProxyQuery + " --> " + promApiResponse.Status)
						}
					}
				}

				if datasource.Type == "Loki" {
					path := "/api/datasources/proxy/" + strconv.Itoa(int(datasource.Id)) + LOKI_QUERYPATH

					res, err = Query(user, passwd, grafanaAdress, path, test.ProxyQuery)
					var lokiResponse map[string]interface{}
					err = json.Unmarshal(res, &lokiResponse)

					if err != nil {
						fmt.Println("Query failed: " + datasource.Name + " --> " + test.ProxyQuery + " --> " + err.Error())
					} else {
						if lokiResponse["streams"] != nil {
							fmt.Println("Query ok: " + datasource.Name + " --> " + test.ProxyQuery)
						} else {
							fmt.Println("Query failed, keine Daten: " + datasource.Name + " --> " + test.ProxyQuery)
						}
					}

				}
			}
		}
	}

}
