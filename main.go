package main

import (
	"bitbucket.org/tebeka/nrsc"
	"bufio"
	"fmt"
	"github.com/ant0ine/go-json-rest"
	"html/template"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	handler := rest.ResourceHandler{}
	handler.SetRoutes(
		rest.Route{"GET", "/api/workers", GetAllWorkers},
		//        rest.Route{"GET", "/workers/:workerName", GetWorker},
	)
	nrsc.Handle("/static/")
	http.HandleFunc("/", Index)
	http.Handle("/api/", &handler)
	http.ListenAndServe(":8080", nil)
}

type GearmanResponse struct {
	Response string
}

func GetInt(s string) (out int) {
	out, err := strconv.Atoi(s)
	if err != nil {
		out = 0
	}
	return
}

func (response GearmanResponse) GetWorkerStatuses() (workerStatuses []GearmanWorkerStatus) {
	statusLines := strings.Split(response.Response, "\n")
	for _, line := range statusLines {
		lineArray := strings.Fields(line)
		workerStatus := GearmanWorkerStatus{
			functionName: lineArray[0],
			jobTotal:     GetInt(lineArray[1]),
			jobRunning:   GetInt(lineArray[2]),
			workerCount:  GetInt(lineArray[3]),
		}
		workerStatuses = append(workerStatuses, workerStatus)
	}
	return
}

type GearmanWorkerStatus struct {
	functionName string
	jobTotal     int
	jobRunning   int
	workerCount  int
}

type GearmanServerInfo struct {
	Address string
	Pid     int
	Version string
}

func GetGearmanServerInfo(address string) GearmanServerInfo {
	pid := SendCommand("getpid", address)
	version := SendCommand("version", address)
	return GearmanServerInfo{address, GetInt(pid.Response), version.Response}
}

type TemplateData struct {
	ServerInfo []GearmanServerInfo
}

func Index(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!doctype html>
<html>
  <head>
    <title>WebGearadmin</title>
    <script src="/static/jquery-2.0.2.min.js"></script>
    <script src="/static/webgearadmin.js"></script>
  </head>
  <body>
    <h3>WebGearadmin</h3>
    <hr>
    {{ range $data := .ServerInfo }}
    Server: {{$data.Address}}<br/>
    PID: {{$data.Pid}}<br/>
    Version: {{$data.Version}}<br/>
    <hr>
    {{ end }}
  </body>
</html>
    `
	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		print(err.Error())
	}
	servers := GetServers(r)
	serverInfo := make([]GearmanServerInfo, len(servers))
	for i, server := range servers {
		serverInfo[i] = GetGearmanServerInfo(server)
	}

	err = t.Execute(w, TemplateData{serverInfo})
	if err != nil {
		print(err.Error())
	}
}

func GetServers(r *http.Request) (servers []string) {
	serversString := r.FormValue("servers")
	if serversString != "" {
		servers = strings.Split(serversString, ",")
	}
	return
}

func SendCommand(cmd, address string) GearmanResponse {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		print(err.Error())
		return GearmanResponse{"error"}
	}
	fmt.Fprint(conn, cmd + "\n")
	response, err := bufio.NewReader(conn).ReadString('\n')
	response = strings.Trim(response, " \n")
	fmt.Printf("RESPONSE FOR |%s|: |%s|\n", cmd, response)
	if cmd == "getpid" || cmd == "version" {
		response = response[3:]
	}
	return GearmanResponse{response}
}

func GetAllWorkers(w *rest.ResponseWriter, r *rest.Request) {
	servers := GetServers(r.Request)
	statuses := make([]GearmanResponse, len(servers))
	for _, server := range servers {
		statusResp := SendCommand("status", server)
		statuses = append(statuses, statusResp)
	}
	w.WriteJson(statuses)
}
