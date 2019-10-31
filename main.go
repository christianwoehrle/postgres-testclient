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

	"github.com/christianwoehrle/prometheus-testclient/dtos"
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
		bodyString := string(bodyBytes)
		fmt.Println(bodyString)
		return bodyBytes, nil

	}
	return nil, fmt.Errorf("error response from server: %d ", resp.StatusCode)

}

func main() {

	filename, _ := filepath.Abs("./check.yaml")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("Fail1")
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

	res, err := Query(user, passwd, "localhost:8080", "/api/datasources", "")

	var datasourcelist dtos.DataSourceList
	err = json.Unmarshal(res, &datasourcelist)

	for i, _ := range datasourceSpecs.Datasource {
		fmt.Println("=================================================================================")
		res, err = Query(user, passwd, "localhost:8080", "/api/datasources/id/"+datasourceSpecs.Datasource[i].Name, "")
		var datasourceId dtos.DataSourceID
		err = json.Unmarshal(res, &datasourceId)
		datasourceSpecs.Datasource[i].Id = datasourceId.Id

	}

	res, err = Query(user, passwd, "localhost:8080", "/api/health", "")
	var health map[string]interface{}
	err = json.Unmarshal(res, &health)
	/*
		{
			"commit": "67bad72",
			"database": "ok",
			"version": "6.3.5"
		}
	*/
	fmt.Println(health["database"])

	for _, datasource := range datasourceSpecs.Datasource {
		fmt.Println("=================================================================================", datasource)
t		for _, test := range datasource.Tests {
			if datasource.Id == 0 {
				fmt.Println("No ID for Datasource <" + datasource.Name + ">, Skip Queries")
			} else {

				if datasource.Type == "Prometheus" {
					path := "/api/datasources/proxy/" + strconv.Itoa(int(datasource.Id)) + PROM_QUERYPATH

					res, err = Query(user, passwd, "localhost:8080", path, test.ProxyQuery)
					fmt.Println(err, string(res))

				}
				if datasource.Type == "Loki" {
					path := "/api/datasources/proxy/" + strconv.Itoa(int(datasource.Id)) + LOKI_QUERYPATH

					res, err = Query(user, passwd, "localhost:8080", path, test.ProxyQuery)
					fmt.Println(err, string(res))

				}
			}
		}
		//TODO Execute Queries
	}

	res, err = Query(user, passwd, "localhost:8080", "/api/datasources/proxy/1/api/v1/query", "query=up{endpoint=\"http-metrics\",instance=\"172.16.248.180:9153\",job=\"coredns\",namespace=\"kube-system\",pod=\"coredns-759d6fc95f-6xq94\",service=\"prometheus-operator-coredns\"}")
	fmt.Println(err)
	var up map[string]interface{}
	err = json.Unmarshal(res, &up)
	/*
		{
			"status": "success",
			"data": {
			"resultType": "vector",
				"result": [
			{
				"metric": {
					"__name__": "up",
					"app": "kube-eagle",
					"instance": "172.16.249.70:8080",
					"job": "kube-eagle",
					"namespace": "monitoring",
					"pod_name": "kube-eagle-6fc6fc7ccf-fhjbf",
					"pod_template_hash": "6fc6fc7ccf",
					"release": "kube-eagle"
				},
				"value": [
					1572449287.973,
				"1"
			]
			},
	*/

	fmt.Println(up["status"])
	if up["status"] == "success" {
		data := up["data"].(map[string]interface{})
		metrics := data["result"].([]interface{})
		fmt.Println(metrics)
		fmt.Printf("%T", metrics)

		for _, result := range metrics {
			fmt.Println(result)
			fmt.Printf("%T", result)
			r := result.(map[string]interface{})
			m := r["metric"]
			v := r["value"]
			fmt.Println(m)
			fmt.Println(v)

		}

	}
	fmt.Println(up)

	fmt.Println("=================================================================================")
	res, err = Query(user, passwd, "localhost:8080", "/api/datasources/proxy/2/api/prom/query", "direction=BACKWARD&limit=1&regexp=&query=%7Bapp%3D%22loki%22%7D")
	fmt.Println(err)
	var loki interface{}
	err = json.Unmarshal(res, &loki)
	/*
		{
		  "streams": [
		    {
		      "labels": "{app=\"loki\", container_name=\"loki\", controller_revision_hash=\"monitoring-loki-54b6787bc4\", filename=\"/var/log/pods/monitoring_monitoring-loki-0_2a93dc3f-fb33-11e9-83c7-022514470a34/loki/0.log\", instance=\"monitoring-loki-0\", job=\"monitoring/loki\", name=\"loki\", namespace=\"monitoring\", release=\"monitoring\", statefulset_kubernetes_io_pod_name=\"monitoring-loki-0\", stream=\"stderr\"}",
		      "entries": [
		        {
		          "ts": "2019-10-30T17:06:49.575248313Z",
		          "line": "level=info ts=2019-10-30T17:06:49.575114026Z caller=table_manager.go:349 msg=\"creating table\" table=index_2595\n"
		        }
		      ]
		    }
		  ]
		}
	*/
	fmt.Println(loki)

}
