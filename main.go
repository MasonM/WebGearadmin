package main

import (
	"bitbucket.org/tebeka/nrsc"
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
		rest.Route{"GET", "/api/:server/workers", GetAllWorkers},
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

func (response GearmanResponse) GetWorkerStatuses() []GearmanWorkerStatus {
	statusLines := strings.Split(response.Response, "\n")
	workerStatuses := make([]GearmanWorkerStatus, len(statusLines))
	for i, line := range statusLines {
		lineArray := strings.Fields(line)
		workerStatuses[i] = GearmanWorkerStatus{
			FunctionName: lineArray[0],
			JobTotal:     GetInt(lineArray[1]),
			JobRunning:   GetInt(lineArray[2]),
			WorkerCount:  GetInt(lineArray[3]),
		}
	}
	return workerStatuses
}

type GearmanWorkerStatus struct {
	FunctionName string
	JobTotal     int
	JobRunning   int
	WorkerCount  int
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
    <h3><a href="https://bitbucket.org/MasonM/webgearadmin">WebGearadmin</a></h3>
    <hr>
    {{ range $data := .ServerInfo }}
    <section id="{{$data.Address}}">
        <h4>Server at {{$data.Address}}</h4>
        PID: {{$data.Pid}}<br/>
        Version: {{$data.Version}}<br/>
        <table name="statuses">
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Total</th>
                    <th>Running</th>
                    <th># Workers</th>
                </tr>
            </thead>
            <tbody></tbody>
        </table>
    </section>
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
	fmt.Fprint(conn, cmd+"\n")

	responseBytes := make([]byte, 1024)
	_, err = conn.Read(responseBytes)
	if err != nil {
		fmt.Printf("Failed to send command \"%s\". Error: %s\n", cmd, err)
	}

	response := strings.Trim(string(responseBytes), " .\r\n\x00")
	if cmd == "getpid" || cmd == "version" {
		// remove leading "OK "
		response = response[3:]
	}
	//fmt.Printf("RESPONSE FOR |%s|: |%s|\n", cmd, response)
	return GearmanResponse{response}
}

func GetAllWorkers(w *rest.ResponseWriter, r *rest.Request) {
	statusResp := SendCommand("status", r.PathParam("server"))
	w.WriteJson(statusResp.GetWorkerStatuses())
}
