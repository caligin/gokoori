package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"regexp"
	"strings"
)

const credentialsFileRelative = ".gokoori/credentials"

//// dashboard unmarshalling types

type Dashboard struct {
	Embedded DashboardEmbedded `json:"_embedded"`
}

type DashboardEmbedded struct {
	PipelineGroups []PipelineGroup `json:"pipeline_groups"`
}

type PipelineGroup struct {
	Name string
	Embedded PipelineGroupEmbedded `json:"_embedded"`
}

type PipelineGroupEmbedded struct {
	Pipelines []Pipeline
}

type Pipeline struct {
	Name string
	Locked bool
	PauseInfo PauseInfo `json:"pause_info"`
	Embedded PipelineEmbedded `json:"_embedded"`
}

type PauseInfo struct {
	Paused bool
	PausedBy string `json:"paused_by"`
	PauseReason string `json:"pause_reason"`
}

type PipelineEmbedded struct {
	Instances []PipelineInstance
}

type PipelineInstance struct {
	Label string
	Embedded PipelineInstanceEmbedded `json:"_embedded"`
}

type PipelineInstanceEmbedded struct {
	Stages []PipelineStage
}

type PipelineStage struct {
	Name string
	Status string
}

//// holder types

type GoApi struct {
	Protocol string
	Host string
	Port uint
}

type Credentials struct {
	Username string
	Password string
}


func pause(client *http.Client, api GoApi, credentials *Credentials, pipeline string, reason string) {
	// TODO make sure reason has the proper encoding? req splitting? (but impact?? like, nope really)
	req, makeReqErr := http.NewRequest("POST", fmt.Sprintf("%s://%s:%d/go/api/pipelines/%s/pause", api.Protocol, api.Host, api.Port, pipeline), strings.NewReader(fmt.Sprintf("pauseCause=%s", reason)))
	if makeReqErr != nil {
		log.Fatalf("error creating pause req for %s: %s", pipeline, makeReqErr.Error())
	}
	req.Header.Add("Confirm", "true")
	if credentials != nil {
		req.SetBasicAuth(credentials.Username, credentials.Password)
	}
	pauseResp, pauseErr := client.Do(req)
	if pauseErr != nil {
		log.Fatalf("error pausing %s: %s", pipeline, pauseErr.Error())
	}
	if pauseResp.StatusCode != 200 {
		log.Fatalf("failed pausing %s, expected 200 got %d", pipeline, pauseResp.StatusCode)
	}
}

func unpause(client *http.Client, api GoApi, credentials *Credentials, pipeline string) {
	req, makeReqErr := http.NewRequest("POST", fmt.Sprintf("%s://%s:%d/go/api/pipelines/%s/unpause", api.Protocol, api.Host, api.Port, pipeline), nil)
	if makeReqErr != nil {
		log.Fatalf("error creating unpause req for %s: %s", pipeline, makeReqErr.Error())
	}
	req.Header.Add("Confirm", "true")
	if credentials != nil {
		req.SetBasicAuth(credentials.Username, credentials.Password)
	}
	unpauseResp, unpauseErr := client.Do(req)
	if unpauseErr != nil {
		log.Fatalf("error unpausing %s: %s", pipeline, unpauseErr.Error())
	}
	if unpauseResp.StatusCode != 200 {
		log.Fatalf("failed unpausing %s, expected 200 got %d", pipeline, unpauseResp.StatusCode)
	}
}

func schedule(client *http.Client, api GoApi, credentials *Credentials, pipeline string) {
	// TODO make sure reason has the proper encoding? req splitting? (but impact?? like, nope really)
	req, makeReqErr := http.NewRequest("POST", fmt.Sprintf("%s://%s:%d/go/api/pipelines/%s/schedule", api.Protocol, api.Host, api.Port, pipeline), nil)
	if makeReqErr != nil {
		log.Fatalf("error creating schedule req for %s: %s", pipeline, makeReqErr.Error())
	}
	req.Header.Add("Confirm", "true")
	if credentials != nil {
		req.SetBasicAuth(credentials.Username, credentials.Password)
	}
	scheduleResp, scheduleErr := client.Do(req)
	if scheduleErr != nil {
		log.Fatalf("error scheduling %s: %s", pipeline, scheduleErr.Error())
	}
	if scheduleResp.StatusCode != 202 {
		log.Fatalf("failed scheduling %s, expected 202 got %d", pipeline, scheduleResp.StatusCode)
	}
}

func isFailed(pipeline *Pipeline) bool {
	if len(pipeline.Embedded.Instances) == 0 {
		return false
	}
	for _, stage := range pipeline.Embedded.Instances[0].Embedded.Stages {
		if stage.Status == "Failed" {
			return true
		}
	}
	return false
}

func dashboard(client *http.Client, api GoApi, credentials *Credentials) Dashboard {
	req, makeReqErr := http.NewRequest("GET", fmt.Sprintf("%s://%s:%d/go/api/dashboard", api.Protocol, api.Host, api.Port), nil)
	if makeReqErr != nil {
		log.Fatalf("error creating list pipeline error: %s", makeReqErr.Error())
	}
	req.Header.Add("Accept", "application/vnd.go.cd.v1+json")
	if credentials != nil {
		req.SetBasicAuth(credentials.Username, credentials.Password)
	}
	dashboardResp, dashboardErr := client.Do(req)
	if dashboardErr != nil {
		log.Fatalf("error getting dashboard: ", dashboardErr.Error())
	}
	if dashboardResp.StatusCode != 200 {
		log.Fatalf("failed retrieving dashboard, expected 200 got %d", dashboardResp.StatusCode)
	}
	defer dashboardResp.Body.Close()
	dashboardBody, dashboardBodyReadErr := ioutil.ReadAll(dashboardResp.Body)
	if dashboardBodyReadErr != nil {
		log.Fatalf("error reading dashboard: %s", dashboardBodyReadErr.Error())
	}
	var dashboard Dashboard
	dashboardUnmarshalErr := json.Unmarshal(dashboardBody, &dashboard)
	if dashboardUnmarshalErr != nil {
		log.Fatalf("error unmarshaling dashboard: %s", dashboardUnmarshalErr.Error())
	}
	return dashboard
}

func readCredentials() *Credentials {
	// TODO bomb if file is world-readable/writable, like ssh for keys.
	user, _ := user.Current()
	homedir := user.HomeDir
	contents, err := ioutil.ReadFile(path.Join(homedir, credentialsFileRelative))
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		log.Fatalf("error opening credentials file", err.Error())
	}
	var credentials Credentials
	jsonErr := json.Unmarshal(contents, &credentials)
	if jsonErr != nil {
		log.Fatalf("error unmarshaling credentials: %s", jsonErr.Error())
	}
	return &credentials
}

func main() {
	//TODO don't pause an already paused as it would ovewrite the pause message? does it even matter?
	//TODO urlencode those url interpolations

	const defaultHttpsPort uint = 8154
	const defaultHttpPort uint = 8153

	var isPause = flag.Bool("pause", false, "Unpause matching pipelines")
	var isUnpause = flag.Bool("unpause", false, "Unpause matching pipelines")
	var isSchedule = flag.Bool("schedule", false, "Schedule matching pipelines")

	var nameFilter = flag.String("name", ".+", "Filter by pipeline name (regex)")
	var groupFilter = flag.String("groupname", ".+", "Filter by group name (regex)")
	var matchFailed = flag.Bool("failed", false, "Match failed pipelines only")
	var matchPaused = flag.Bool("paused", false, "Match paused pipelines only")
	var reason = flag.String("reason", "Paused by Gokoori", "Reason for pausing")
	var host = flag.String("host", "localhost", "Host hunning the go server")
	var port = flag.Uint("port", defaultHttpsPort, "Port the go server is running on")
	var insecure = flag.Bool("insecure", false, "Use HTTP plain instead of HTTPS")

	flag.Parse()
	if *isPause && *isUnpause {
		// TODO and how does trigger interact with this???? pause and trigger is not sound but unpause and schedule kinda is
		log.Fatalf("Both pause and unpause are specified, those two options cannot be used together")
	}
	if *insecure && (*port == defaultHttpsPort) {
		// TODO this means that if someone specifies the defaultHttpsPort explicitly we can't yet detect it and just override it with defaultHttpPort...
		*port = defaultHttpPort
	}
	var protocol = "https"
	if *insecure {
		protocol = "http"
	}
	api := GoApi{protocol, *host, *port}

	client := &http.Client{}

	credentials := readCredentials()

	dashboard := dashboard(client, api, credentials)

	for _, group := range dashboard.Embedded.PipelineGroups {
		match, matchErr := regexp.MatchString(*groupFilter, group.Name)
		if matchErr != nil {
			log.Fatalf("error regexing %s: %s", group.Name, matchErr)
		}
		if !match {
			continue
		}
		for _, pipeline := range group.Embedded.Pipelines {
			match, matchErr := regexp.MatchString(*nameFilter, pipeline.Name)
			if matchErr != nil {
				log.Fatalf("error regexing %s: %s", pipeline.Name, matchErr)
			}
			if !match {
				continue
			}
			if *matchFailed && !isFailed(&pipeline) {
				continue
			}
			if *matchPaused && !pipeline.PauseInfo.Paused {
				continue
			}
			fmt.Println(pipeline.Name)
			if *isUnpause {
				unpause(client, api, credentials, pipeline.Name)
			}
			if *isPause {
				pause(client, api, credentials, pipeline.Name, *reason)
			}
			if *isSchedule {
				schedule(client, api, credentials, pipeline.Name)
			}
		}
	}

}
