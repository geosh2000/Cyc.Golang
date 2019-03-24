package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func getAPI(tp string, route string, uri string, params url.Values) (result string, err error) {

	pTypes := []string{
		"POST",
		"GET",
	}

	tF, _ := inArray(tp, pTypes)

	//Get data from QM (API)
	if route == "" {
		err = fmt.Errorf("La ruta no debe ser nula")
		return
	}
	if uri == "" {
		err = fmt.Errorf("La uri no debe ser nula")
		return
	}
	if !tF {
		err = fmt.Errorf("Ingresa un tipo de REQUEST correcto")
		return
	}

	var req *http.Request
	adr := fmt.Sprintf("%s%s", uri, route)
	if tp == "GET" {
		req, err = http.NewRequest(tp, adr, nil)
	} else {
		req, err = http.NewRequest(tp, adr, bytes.NewBufferString(params.Encode()))
	}

	req.Header.Set("content-type", `application/x-www-form-urlencoded; param=value`)

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

	if posts["ERR"].(bool) {
		err = fmt.Errorf(posts["error"].(string))
	}

	result = posts["msg"].(string)

	return
}
