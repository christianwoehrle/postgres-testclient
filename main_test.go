package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func Test_ParseDatasource(t *testing.T) {
	input := []byte(`datasources:
- name: Loki
  type: Loki
  tests:
    - proxyQuery: "direction=BACKWARD&limit=1&regexp=&query=%7Bapp%3D%22loki%22%7D"
- name: Prometheus
  type: Prometheus
  tests:
    - proxyQuery: "query=up{endpoint=\"http-metrics\",instance=\"172.16.248.180:9153\",job=\"coredns\",namespace=\"kube-system\",pod=\"coredns-759d6fc95f-6xq94\",service=\"prometheus-operator-coredns\"}"
`)

	datasources := Datasources{
		[]Datasource{
			{
				Name: "Loki",
				Type: "Loki",

				Tests: []Test{
					{
						ProxyQuery: "direction=BACKWARD&limit=1&regexp=&query=%7Bapp%3D%22loki%22%7D",
					},
				},
			},
			{
				Name: "Prometheus",
				Type: "Prometheus",

				Tests: []Test{
					{
						ProxyQuery: "query=up{endpoint=\"http-metrics\",instance=\"172.16.248.180:9153\",job=\"coredns\",namespace=\"kube-system\",pod=\"coredns-759d6fc95f-6xq94\",service=\"prometheus-operator-coredns\"}",
					},
				},
			},
		},
	}

	actual, err := ParseDatasources(input)
	require.NoError(t, err)
	require.NotNil(t, actual)
	require.Len(t, actual.Datasource, 2)
	assert.Equal(t, datasources, actual)

}

func ParseDatasources(yamlFile []byte) (Datasources, error) {
	var f Datasources
	err := yaml.Unmarshal(yamlFile, &f)
	return f, err
}
