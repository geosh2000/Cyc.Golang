package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func runAgentCalls(paises []string) {

	fmt.Printf("======== Agents Live START ========\n")

	// Conexión a bases de datos
	// dbXp, err := sql.Open("mysql", dbCon[1])
	// if err != nil {
	// 	panic(err.Error())
	// }

	for _, pais := range paises {

		fmt.Printf("|- %s", pais)

		base := "liveMonitor"
		if pais == "MX" {
			base += "MX"
		}
		updtFlagQ := fmt.Sprintf("UPDATE %s SET updateFlag = %d", base, 0)
		r, err := fDbXp.Exec(updtFlagQ)
		if err != nil {
			fmt.Printf("Error al accesar a base de datos %s\n", pais)
			fmt.Printf(". Error: %s\n", err.Error())
			return
		}
		_ = r
		fmt.Printf(".")

		query, err := qmAgents(pais, base)
		if err != nil {
			fmt.Printf("Error al accesar a base de datos %s\n", pais)
			fmt.Printf(". Error: %s\n", err.Error())
			return
		}
		fmt.Printf(".")

		q, err := fDbXp.Exec(query)
		if err != nil {
			fmt.Printf("Error al actualizar liveAgents %s\n", pais)
			fmt.Printf(". Error: %s\n", err.Error())
			return
		}
		_ = q
		fmt.Printf(". %5s\n", "OK!")

		del := fmt.Sprintf("DELETE FROM %s WHERE updateFlag = 0 AND Agent NOT LIKE 'wait%s'", base, "%")
		d, err := fDbXp.Exec(del)
		if err != nil {
			fmt.Printf("Error al borrar agentes inactivos %s\n", pais)
			fmt.Printf(". Error: %s\n", err.Error())
			return
		}
		_ = d
		fmt.Printf(". %5s\n", "OK!")
	}

	fmt.Printf("======== Agents Live   END ========\n")

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

	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(data.Encode()))
	req.Header.Set("content-type", `application/x-www-form-urlencoded; param=value`)
	req.Header.Add("Authorization", `Basic cm9ib3Q6cm9ib3Q=`)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	//Convert response to readable array
	body, _ := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var posts map[string]interface{}
	json.Unmarshal(body, &posts)

	// Get Fields
	var fields []string
	fNames := "updateFlag"
	for _, v := range posts[block].([]interface{})[0].([]interface{}) {
		fields = append(fields, v.(string)[prefix:])

		fNames += ","
		fNames += v.(string)[prefix:]
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
