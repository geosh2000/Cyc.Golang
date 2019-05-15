package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func runAgentCalls(paises []string) {

	tl := "Agents Live"

	printFrame(tl, true)
	formatPrint("", false)

	// Conexión a bases de datos
	// dbXp, err := sql.Open("mysql", dbCon[1])
	// if err != nil {
	// 	panic(err.Error())
	// }

	for _, pais := range paises {

		base := "liveMonitor"
		if pais == "MX" {
			base += "MX"
		}

		formatPrint(fmt.Sprintf("Build query for insert %s", pais), true)
		query, err := qmAgents(pais, base)
		_ = query
		printStatus("", err)

		if err != nil {
			continue
		}

		formatPrint(fmt.Sprintf("Flag update to 1 %s", pais), true)
		updtFlagQ := fmt.Sprintf("UPDATE %s SET updateFlag = %d", base, 0)
		r, err := fDbXp.Exec(updtFlagQ)
		_ = r
		printStatus("", err)

		formatPrint(fmt.Sprintf("Insert to DB %s", pais), true)
		q, err := fDbXp.Exec(query)
		_ = q
		printStatus("", err)

		formatPrint(fmt.Sprintf("Delete inactive agents %s", pais), true)
		del := fmt.Sprintf("DELETE FROM %s WHERE updateFlag = 0 AND Agent NOT LIKE 'wait%s'", base, "%")
		d, err := fDbXp.Exec(del)
		_ = d
		printStatus("", err)
	}

	formatPrint("", false)
	printFrame(tl, false)

}

func qmAgents(pais string, b string) (query string, err error) {

	//Get data from QM (API)
	block := "RealTimeDO.RtAgentsRaw"
	prefix := 4
	data := url.Values{}
	uri := uriMX[0]
	data.Set("queues", "*")
	data.Add("block", block)

	if pais == "CO" {
		uri = uriCO[0]
	}

	body, fields, _, err := getFromQm("POST", uri, block, prefix, true, data)
	if err != nil {
		return
	}

	var posts map[string]interface{}

	if body == nil {
		err = fmt.Errorf("Información Nula")
		return
	}

	json.Unmarshal(body, &posts)

	// Get Fields

	fNames := "updateFlag"
	for _, v := range posts[block].([]interface{})[0].([]interface{}) {
		fNames += ","
		fNames += v.(string)[prefix:]
	}

	if len(posts[block].([]interface{})[1:]) == 0 {
		err = fmt.Errorf("Sin Data para insertar")
		return
	}

	var values []string
	query = fmt.Sprintf("INSERT INTO %s (%s) VALUES ", b, fNames)
	for i, v := range posts[block].([]interface{})[1:] {
		valor := "(1"
		for _, r := range v.([]interface{}) {
			valor += ","
			valor += "'" + strings.Trim(strings.Replace(r.(string), "&nbsp;", "", -1), " ") + "'"
		}
		valor += ")"
		if i != 0 {
			query += ", "
		}
		query += fmt.Sprintf("%s", valor)
		values = append(values, valor)
	}

	update := "updateFlag = 1"
	for _, v := range fields {
		update += fmt.Sprintf(",%s=VALUES(%s)", v, v)
	}

	query += fmt.Sprintf("ON DUPLICATE KEY UPDATE %s;", update)
	return
}
