package main

import (
    "fmt"
    "net/http"
    "strings"
    "encoding/json"
    "io/ioutil"
    "regexp"
)

type Pipeline struct {
  Name string
  // and other things ignored for now
}

type PipelineGroup struct {
  Name string
  Pipelines []Pipeline
}

func pause(client *http.Client, pipeline string) {
    req, makeReqErr := http.NewRequest("POST", fmt.Sprintf("http://localhost:32771/go/api/pipelines/%s/pause", pipeline), strings.NewReader("pauseCause=b/c I can"))
    if makeReqErr != nil {
        panic(fmt.Sprintf("error creating pause req for %s: %s", pipeline, makeReqErr.Error()))
    }
    // handle error ofc
    req.Header.Add("Confirm","true")
    pauseResp, pauseErr := client.Do(req)
    if pauseErr != nil {
        panic(fmt.Sprintf("error pausing %s: %s", pipeline, pauseErr.Error()))
    }
    if pauseResp.StatusCode != 200 {
        panic(fmt.Sprintf("failed pausing %s, expected 200 got %s", pipeline, pauseResp.StatusCode))
    }
}

func main() {
// curl -X POST 'localhost:32771/go/api/pipelines/asd-training/pause' -d 'pauseCause=b/c I can' -H 'Confirm: true' -vvv
// https://go-server-url:8154/go/api
// application/vnd.go.cd.v1+json
// POST /go/api/pipelines/:pipeline_name/pause
// pauseCause

//    http.Post("localhost:32771/go/api/pipelines/asd-training/pause")
    client := &http.Client{}

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

    for _, group := range pipelineGroups {
        for _, pipeline := range group.Pipelines {
            match, matchErr := regexp.MatchString("training", pipeline.Name)
            if matchErr != nil {
                panic(fmt.Sprintf("error regexing %s: %s", pipeline.Name, pgBodyReadErr.Error()))
            }
            if match {
                pause(client, pipeline.Name)
            }
        }
    }


}
