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

func pause(client *http.Client, pipeline string, reason string) {
	// TODO make sure reason has the proper encoding? req splitting? (but impact?? like, nope really)
	req, makeReqErr := http.NewRequest("POST", fmt.Sprintf("http://localhost:32771/go/api/pipelines/%s/pause", pipeline), strings.NewReader(fmt.Sprintf("pauseCause=%s", reason)))
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

func unpause(client *http.Client, pipeline string) {
	req, makeReqErr := http.NewRequest("POST", fmt.Sprintf("http://localhost:32771/go/api/pipelines/%s/unpause", pipeline), nil)
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

func listGroups(client *http.Client) []PipelineGroup {
	pgResp, pgErr := client.Get("http://localhost:32771/go/api/config/pipeline_groups")
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
	flag.Parse()
	if *isPause && *isUnpause {
		panic("Both pause and unpause are specified, those two options cannot be used together")
	}

	client := &http.Client{}

	pipelineGroups := listGroups(client)

	for _, group := range pipelineGroups {
		for _, pipeline := range group.Pipelines {
			match, matchErr := regexp.MatchString(*nameFilter, pipeline.Name)
			if matchErr != nil {
				panic(fmt.Sprintf("error regexing %s: %s", pipeline.Name, matchErr))
			}
			if match {
				fmt.Println(pipeline.Name)
				if *isUnpause {
					unpause(client, pipeline.Name)
				}
				if *isPause {
					pause(client, pipeline.Name, *reason)
				}
			}
		}
	}

}
