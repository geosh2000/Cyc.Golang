package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func runLiveCalls(paises []string) {

	fmt.Printf("======== Calls  Live START ========\n")

	// // Conexión a bases de datos
	// db, err := sql.Open("mysql", dbCon[0])
	// if err != nil {
	// 	panic(err.Error())
	// }
	// dbXp, err := sql.Open("mysql", dbCon[1])
	// if err != nil {
	// 	panic(err.Error())
	// }

	fmt.Printf("|- Obteniendo colas OUT.")

	qOut, err := fDb.Query("SELECT queue FROM Cola_Skill WHERE direction=2")
	if err != nil {
		fmt.Printf("Error al obtener colas Outbound\n")
		fmt.Printf("|- Error: %s\n", err.Error())
		return
	}
	fmt.Printf(".")
	defer qOut.Close()

	var outQs []string
	for qOut.Next() {
		var queue string
		// for each row, scan the result into our tag composite object
		err = qOut.Scan(&queue)
		if err != nil {
			fmt.Println(err.Error())
		}
		outQs = append(outQs, queue)
	}
	fmt.Printf(". OK!\n")

	for _, pais := range paises {

		fmt.Printf("|- %s", pais)

		base := "liveMonitor"
		if pais == "MX" {
			base += "MX"
		}
		updtFlagQ := fmt.Sprintf("UPDATE %s SET updateFlag = %d", base, 0)
		r, err := fDbXp.Exec(updtFlagQ)
		if err != nil {
			fmt.Printf("Error al cambiar el status del flag %s\n", pais)
			fmt.Printf("|- Error: %s\n", err.Error())
			return
		}
		_ = r
		fmt.Printf(".")

		query, qTn, err := qmCalls(pais, outQs, base)
		if err != nil {
			fmt.Printf("Error al correr llamadas en vivo %s\n", pais)
			fmt.Printf("|- Error: %s\n", err.Error())
			return
		}
		fmt.Printf(".")

		_ = query
		q, err := fDbXp.Exec(query)
		if err != nil {
			fmt.Printf("Error al actualizar liveCalls %s\n", pais)
			fmt.Printf(". Error: %s\n", err.Error())
			return
		}
		_ = q
		fmt.Printf(".")

		updtFlagQ = fmt.Sprintf(qTn)
		ru, err := fDbXp.Exec(updtFlagQ)
		if err != nil {
			fmt.Printf("Error al setear fields en NULL %s\n", pais)
			fmt.Printf("|- Error: %s\n", err.Error())
			// fmt.Printf("%s\n", qTn)
			return
		}
		_ = ru
		fmt.Printf(". %5s\n", "OK!")
	}

	fmt.Printf("======== Calls  Live   END ========\n")

}

func qmCalls(pais string, outQs []string, b string) (query string, queryToNull string, err error) {
	//Get data from QM (API)
	block := "RealTimeDO.RtCallsRaw"
	prefix := 0
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
	fNames := "updateFlag,Agent"
	for _, v := range posts[block].([]interface{})[0].([]interface{}) {
		fields = append(fields, v.(string)[prefix:])

		fNames += ","
		fNames += v.(string)[prefix:]
	}

	// Get important field indexes
	_, rtQ := inArray("RT_queue", fields)
	_, rtU := inArray("RT_url", fields)
	_, rtA := inArray("RT_answered", fields)
	_, rtAg := inArray("RT_agent", fields)
	_, rtSi := inArray("RT_serverId", fields)

	// Create Array for sorting
	var vals [][]string
	for _, v := range posts[block].([]interface{})[1:] {
		var thisVals []string
		for x, r := range v.([]interface{}) {

			// Replace URL for outbound flag
			if x == rtU {
				f, _ := inArray(v.([]interface{})[rtQ], outQs)
				if f {
					thisVals = append(thisVals, "1")
				} else {
					thisVals = append(thisVals, "0")
				}
			} else {
				thisVals = append(thisVals, strings.Trim(strings.Replace(r.(string), "&nbsp;", "", -1), " "))
			}
		}
		vals = append(vals, thisVals)
	}

	// Sort by Answered
	sort.Slice(vals, func(i, j int) bool {
		return vals[i][rtA] < vals[j][rtA]
	})

	// Build query
	var values []string
	item := 1
	query = fmt.Sprintf("INSERT INTO %s (%s) VALUES ", b, fNames)
	for i, v := range vals {
		valor := "(1"

		//Agent replace for primary key
		if v[rtAg] == "" {
			valor += fmt.Sprintf(",'wait%d'", item)
			item++
		} else {
			valor += fmt.Sprintf(",'%s'", v[rtAg])
		}

		for _, r := range v {
			valor += ","
			valor += "'" + strings.Trim(strings.Replace(r, "&nbsp;", "", -1), " ") + "'"
		}
		valor += ")"
		if i != 0 {
			query += ", "
		}
		query += fmt.Sprintf("%s", valor)
		values = append(values, valor)
	}

	update := "updateFlag = 1"
	updToNull := "RT_queue=NULL"
	for x, v := range fields {

		switch pais {
		case "CO":
			// Para mostrar en monitor todas las llamadas salientes
			update += fmt.Sprintf(",%s=VALUES(%s)", v, v)

			// Fallthroug when last option not active
			// fallthrough
		case "MX":
			// Para mostrar en monitor solo llamadas salientes cuando no tiene llamada entrante
			if x == rtSi {
				update += ",obCaller=IF(VALUES(RT_url)=1, VALUES(RT_Caller), NULL), obTst=IF(VALUES(RT_url)=1, VALUES(RT_answered), NULL)"
			} else {
				update += fmt.Sprintf(",%s=IF(updateFlag=1 AND RT_url=0 AND VALUES(RT_url)=1, %s, VALUES(%s))", v, v, v)
			}
		}

		updToNull += fmt.Sprintf(",%s=NULL", v)
	}

	if pais == "CO" {
		update += ",Freesincepauorcalltst = UNIX_TIMESTAMP(NOW())"
	} else {
		updToNull = fmt.Sprintf("obCaller = NULL, obTst= NULL, %s", updToNull)
	}

	query += fmt.Sprintf("ON DUPLICATE KEY UPDATE %s;", update)
	queryToNull = fmt.Sprintf("UPDATE %s SET %s WHERE updateFlag = 0", b, updToNull)
	return
}
