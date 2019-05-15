package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func getPauses() {

	sp.Stop()
	tl := "Agent Pauses"

	printFrame(tl, true)
	formatPrint("", false)

	paises := valPaises()

	var err error
	for _, p := range paises {
		err = runPauses(p)
	}

	formatPrint("Recalculating Pauses", true)
	sp.Start()
	//Get data from QM (API)
	var req *http.Request
	adr := "http://testoperaciones.pricetravel.com.mx/api/restfulbck/index.php/Procesos/pauseCheck"
	req, err = http.NewRequest("GET", adr, nil)

	req.Header.Set("content-type", `application/x-www-form-urlencoded; param=value`)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}
	sp.Stop()
	printStatus("Pausas Correctamente recalculadas", err)

	formatPrint("", false)
	printFrame(tl, false)

	flag := prReload(5)
	if flag {
		getPauses()
	}

}

func runPauses(pais string) (err error) {
	block := "DetailsDO.AgentPauses"
	uri := uriMX[1]

	if pais == "CO" {
		uri = uriCO[1]
	}

	td := time.Now()
	tm := td.AddDate(0, 0, 1)

	data := url.Values{}
	data.Set("queues", "*")
	data.Add("block", block)
	data.Add("from", fmt.Sprintf("%d-%02d-%02d.00:00:00", td.Year(), td.Month(), td.Day()))
	data.Add("to", fmt.Sprintf("%d-%02d-%02d.00:00:00", tm.Year(), tm.Month(), tm.Day()))

	formatPrint(fmt.Sprintf("Getting Pauses %s", pais), true)

	sp.Start()
	body, fields, values, err := getFromQm("POST", uri, block, 0, true, data)
	_, _ = body, fields
	sp.Stop()
	printStatus("", err)
	if err != nil {
		return
	}

	var posts map[string]interface{}
	json.Unmarshal(body, &posts)

	var vQ string
	for i, v := range values {

		// Asesor
		ras := regexp.MustCompile(`\([0-9]+\)`)
		asesor := fmt.Sprintf("GETIDASESOR('%s', 2)", strings.TrimSpace(ras.ReplaceAllString(v[0], "")))

		// Code
		code := v[2]
		if code == "-" {
			code = "0"
		}

		// Inicio y Fin
		year := time.Now().Year()
		rDt := regexp.MustCompile(`/`)
		rDT := regexp.MustCompile(` - `)

		inVal := fmt.Sprintf("%d-%s-00:00", year, rDT.ReplaceAllString(rDt.ReplaceAllString(v[6], "-"), "T"))
		outVal := fmt.Sprintf("%d-%s-00:00", year, rDT.ReplaceAllString(rDt.ReplaceAllString(v[7], "-"), "T"))

		dt, _ := time.Parse(time.RFC3339, inVal)
		inicio := fmt.Sprintf("'%d-%02d-%02d %02d:%02d:%02d'", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second())
		dt, _ = time.Parse(time.RFC3339, outVal)
		fin := fmt.Sprintf("'%d-%02d-%02d %02d:%02d:%02d'", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second())

		// Duracion
		duracion := fmt.Sprintf("TIMEDIFF(%s,%s)", fin, inicio)

		// Q Value
		qVal := fmt.Sprintf("(%s,%s,TZCONVERT(%s),TZCONVERT(%s),%s,%d)", asesor, code, inicio, fin, duracion, 0)

		if i > 0 {
			vQ += ","
		}

		vQ += qVal
	}

	if vQ == "" {
		err = fmt.Errorf("Sin Resultados para insertar")
		printStatus("", err)
		return
	}

	formatPrint(fmt.Sprintf("Inserting Pauses %s", pais), true)
	sp.Start()
	query := fmt.Sprintf("INSERT INTO asesores_pausas (asesor,tipo,Inicio,Fin,Duracion,Skill) VALUES %s ON DUPLICATE KEY UPDATE Fin=VALUES(Fin), Duracion=VALUES(Duracion)", vQ)
	sq, err := fDb.Exec(query)
	_ = sq
	sp.Stop()
	printStatus("", err)

	if err != nil {
		fmt.Println(query)
	}

	return
}
