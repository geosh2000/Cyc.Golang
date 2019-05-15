package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
)

func runAgDetails() {
	sp.Stop()
	paises := []string{"MX", "CO"}

	tl := "Agent Details"

	printFrame(tl, true)
	formatPrint("", false)

	for _, pais := range paises {

		query, err := agDetails(pais)
		printStatus("", err)

		formatPrint(fmt.Sprintf("Updating AgentDetails Table %s", pais), true)
		sp.Start()
		r, err := fDbXp.Exec(query)
		sp.Stop()
		_ = r
		printStatus("", err)
	}

	formatPrint("", false)
	printFrame(tl, false)
}

func agDetails(pais string) (query string, err error) {

	sp.Stop()
	//Get data from QM (API)
	data := url.Values{}
	uri := "http://queuemetrics.pricetravel.com.mx:8080/queuemetricscc/agent/jsonEditorApi.do"

	if pais == "CO" {
		uri = "http://queuemetrics-co.pricetravel.com.mx:8080/qm/agent/jsonEditorApi.do"
	}

	formatPrint(fmt.Sprintf("Getting Agents data %s", pais), true)
	sp.Start()
	body, _, _, err := getFromQm("GET", uri, "", 0, false, data)
	sp.Stop()
	printStatus("", err)

	formatPrint(fmt.Sprintf("Building Query %s", pais), true)

	type Agents struct {
		Nome_agente      string
		Group_name       string
		Real_name        string
		Current_terminal string
		Vnc_url          string
		Supervised_by    string
		Descr_agente     string
		Xmpp_address     string
		PK_ID            string
		Group_icon       string
		Group_by         string
		Chiave_agente    string
		Location         string
		Loc_name         string
	}

	var agents []Agents
	json.Unmarshal(body, &agents)

	l := reflect.ValueOf(agents[0])
	typeOfT := l.Type()
	// fields := make(map[int]string, l.NumField())

	var fields []string
	var qU string
	qF := "("
	for x := 0; x < l.NumField(); x++ {
		// fields = append(fields, l.Field(x).Interface().(string))
		fields = append(fields, typeOfT.Field(x).Name)
		if x > 0 {
			qF += ","
			qU += ","
		}
		qF += typeOfT.Field(x).Name
		qU += fmt.Sprintf("%[1]s=VALUES(%[1]s)", typeOfT.Field(x).Name)
	}
	qF += ")"

	var qV string
	for i, v := range agents {
		l = reflect.ValueOf(v)
		var tqV, sep string
		for x := 0; x < l.NumField(); x++ {
			if x > 0 {
				tqV += ","
			}
			tqV += fmt.Sprintf("'%s'", l.Field(x).Interface().(string))
		}
		if i > 0 {
			sep += ","
		}
		qV += fmt.Sprintf("%s (%s)", sep, tqV)

	}

	query = fmt.Sprintf("INSERT INTO ccexporter.agentDetails %s VALUES %s ON DUPLICATE KEY UPDATE %s", qF, qV, qU)

	return
}
