package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
	"regexp"
	"strings"
)

const credentialsFileRelative = ".gokoori/credentials"

type Pipeline struct {
	Name string
	// and other things ignored for now
}

type PipelineGroup struct {
	Name      string
	Pipelines []Pipeline
}

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
		panic(fmt.Sprintf("error creating pause req for %s: %s", pipeline, makeReqErr.Error()))
	}
	req.Header.Add("Confirm", "true")
	if credentials != nil {
		req.SetBasicAuth(credentials.Username, credentials.Password)
	}
	pauseResp, pauseErr := client.Do(req)
	if pauseErr != nil {
		panic(fmt.Sprintf("error pausing %s: %s", pipeline, pauseErr.Error()))
	}
	if pauseResp.StatusCode != 200 {
		panic(fmt.Sprintf("failed pausing %s, expected 200 got %d", pipeline, pauseResp.StatusCode))
	}
}

func unpause(client *http.Client, api GoApi, credentials *Credentials, pipeline string) {
	req, makeReqErr := http.NewRequest("POST", fmt.Sprintf("%s://%s:%d/go/api/pipelines/%s/unpause", api.Protocol, api.Host, api.Port, pipeline), nil)
	if makeReqErr != nil {
		panic(fmt.Sprintf("error creating unpause req for %s: %s", pipeline, makeReqErr.Error()))
	}
	req.Header.Add("Confirm", "true")
	if credentials != nil {
		req.SetBasicAuth(credentials.Username, credentials.Password)
	}
	unpauseResp, unpauseErr := client.Do(req)
	if unpauseErr != nil {
		panic(fmt.Sprintf("error unpausing %s: %s", pipeline, unpauseErr.Error()))
	}
	if unpauseResp.StatusCode != 200 {
		panic(fmt.Sprintf("failed unpausing %s, expected 200 got %d", pipeline, unpauseResp.StatusCode))
	}
}

func listGroups(client *http.Client, api GoApi, credentials *Credentials) []PipelineGroup {
	req, makeReqErr := http.NewRequest("GET", fmt.Sprintf("%s://%s:%d/go/api/config/pipeline_groups", api.Protocol, api.Host, api.Port), nil)
	if makeReqErr != nil {
		panic(fmt.Sprintf("error creating list pipeline error: %s", makeReqErr.Error()))
	}
	if credentials != nil {
		req.SetBasicAuth(credentials.Username, credentials.Password)
	}
	pgResp, pgErr := client.Do(req)
	if pgErr != nil {
		panic(fmt.Sprintf("error getting pipeline groups: %s", pgErr.Error()))
	}
	if pgResp.StatusCode != 200 {
		panic(fmt.Sprintf("failed listing pipeline groups, expected 200 got %d", pgResp.StatusCode))
	}
	defer pgResp.Body.Close()
	pgBody, pgBodyReadErr := ioutil.ReadAll(pgResp.Body)
	if pgBodyReadErr != nil {
		panic(fmt.Sprintf("error reading pipeline groups: %s", pgBodyReadErr.Error()))
	}
	var pipelineGroups []PipelineGroup
	pgUnmarshalErr := json.Unmarshal(pgBody, &pipelineGroups)
	if pgUnmarshalErr != nil {
		panic(fmt.Sprintf("error unmarshaling pipeline groups: %s", pgUnmarshalErr.Error()))
	}
	return pipelineGroups
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
		panic(fmt.Sprintf("error opening credentials file", err.Error()))
	}
	var credentials Credentials
	jsonErr := json.Unmarshal(contents, &credentials)
	if jsonErr != nil {
		panic(fmt.Sprintf("error unmarshaling credentials: %s", jsonErr.Error()))
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

	var nameFilter = flag.String("name", ".+", "Filter by pipeline name (regex)")
	var reason = flag.String("reason", "Paused by Gokoori", "Reason for pausing")
	var host = flag.String("host", "localhost", "Host hunning the go server")
	var port = flag.Uint("port", defaultHttpsPort, "Port the go server is running on")
	var insecure = flag.Bool("insecure", false, "Use HTTP plain instead of HTTPS")

	flag.Parse()
	if *isPause && *isUnpause {
		panic("Both pause and unpause are specified, those two options cannot be used together")
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

	pipelineGroups := listGroups(client, api, credentials)

	for _, group := range pipelineGroups {
		for _, pipeline := range group.Pipelines {
			match, matchErr := regexp.MatchString(*nameFilter, pipeline.Name)
			if matchErr != nil {
				panic(fmt.Sprintf("error regexing %s: %s", pipeline.Name, matchErr))
			}
			if match {
				fmt.Println(pipeline.Name)
				if *isUnpause {
					unpause(client, api, credentials, pipeline.Name)
				}
				if *isPause {
					pause(client, api, credentials, pipeline.Name, *reason)
				}
			}
		}
	}

}
