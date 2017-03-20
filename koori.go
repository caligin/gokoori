package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type Pipeline struct {
	Name string
	// and other things ignored for now
}

type PipelineGroup struct {
	Name      string
	Pipelines []Pipeline
}

type GoApi struct {
	Host string
	Port uint
}

func pause(client *http.Client, api GoApi, pipeline string, reason string) {
	// TODO make sure reason has the proper encoding? req splitting? (but impact?? like, nope really)
	req, makeReqErr := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/go/api/pipelines/%s/pause", api.Host, api.Port, pipeline), strings.NewReader(fmt.Sprintf("pauseCause=%s", reason)))
	if makeReqErr != nil {
		panic(fmt.Sprintf("error creating pause req for %s: %s", pipeline, makeReqErr.Error()))
	}
	req.Header.Add("Confirm", "true")
	pauseResp, pauseErr := client.Do(req)
	if pauseErr != nil {
		panic(fmt.Sprintf("error pausing %s: %s", pipeline, pauseErr.Error()))
	}
	if pauseResp.StatusCode != 200 {
		panic(fmt.Sprintf("failed pausing %s, expected 200 got %s", pipeline, pauseResp.StatusCode))
	}
}

func unpause(client *http.Client, api GoApi, pipeline string) {
	req, makeReqErr := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/go/api/pipelines/%s/unpause", api.Host, api.Port, pipeline), nil)
	if makeReqErr != nil {
		panic(fmt.Sprintf("error creating unpause req for %s: %s", pipeline, makeReqErr.Error()))
	}
	req.Header.Add("Confirm", "true")
	unpauseResp, unpauseErr := client.Do(req)
	if unpauseErr != nil {
		panic(fmt.Sprintf("error unpausing %s: %s", pipeline, unpauseErr.Error()))
	}
	if unpauseResp.StatusCode != 200 {
		panic(fmt.Sprintf("failed unpausing %s, expected 200 got %s", pipeline, unpauseResp.StatusCode))
	}
}

func listGroups(client *http.Client, api GoApi) []PipelineGroup {
	pgResp, pgErr := client.Get(fmt.Sprintf("http://%s:%d/go/api/config/pipeline_groups", api.Host, api.Port))
	if pgErr != nil {
		panic(fmt.Sprintf("error getting pipeline groups: %s", pgErr.Error()))
	}
	defer pgResp.Body.Close()
	pgBody, pgBodyReadErr := ioutil.ReadAll(pgResp.Body)
	if pgBodyReadErr != nil {
		panic(fmt.Sprintf("error reading pipeline groups: %s", pgBodyReadErr.Error()))
	}
	var pipelineGroups []PipelineGroup
	pgUnmarshalErr := json.Unmarshal(pgBody, &pipelineGroups)
	if pgUnmarshalErr != nil {
		panic(fmt.Sprintf("error unmarshaling pipeline groups: %s", pgBodyReadErr.Error()))
	}
	return pipelineGroups
}

func main() {
	//TODO don't pause an already paused as it would ovewrite the pause message? does it even matter?
	//TODO urlencode those url interpolations

	var isPause = flag.Bool("pause", false, "Unpause matching pipelines")
	var isUnpause = flag.Bool("unpause", false, "Unpause matching pipelines")
	// TODO quiet flag ?
	var nameFilter = flag.String("name", ".+", "Filter by pipeline name (regex)")
	var reason = flag.String("reason", "Paused by Gokoori", "Reason for pausing")
	var host = flag.String("host", "localhost", "Host hunning the go server")
	var port = flag.Uint("port", 8153, "Port the go server is running on")
	flag.Parse()
	if *isPause && *isUnpause {
		panic("Both pause and unpause are specified, those two options cannot be used together")
	}
	api := GoApi{*host, *port}

	client := &http.Client{}

	pipelineGroups := listGroups(client, api)

	for _, group := range pipelineGroups {
		for _, pipeline := range group.Pipelines {
			match, matchErr := regexp.MatchString(*nameFilter, pipeline.Name)
			if matchErr != nil {
				panic(fmt.Sprintf("error regexing %s: %s", pipeline.Name, matchErr))
			}
			if match {
				fmt.Println(pipeline.Name)
				if *isUnpause {
					unpause(client, api, pipeline.Name)
				}
				if *isPause {
					pause(client, api, pipeline.Name, *reason)
				}
			}
		}
	}

}
