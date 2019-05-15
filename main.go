package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/briandowns/spinner"
	"github.com/eidolon/wordwrap"
	_ "github.com/go-sql-driver/mysql"
)

//QM address
var uriMX = [2]string{
	"http://queuemetrics.pricetravel.com.mx:8080/queuemetricscc/QmRealtime/jsonStatsApi.do",
	"http://queuemetrics.pricetravel.com.mx:8080/queuemetricscc/QmStats/jsonStatsApi.do",
}
var uriCO = [2]string{
	"http://queuemetrics-co.pricetravel.com.mx:8080/qm/QmRealtime/jsonStatsApi.do",
	"http://queuemetrics-co.pricetravel.com.mx:8080/qm/QmStats/jsonStatsApi.do",
}
var dbCon = [2]string{
	"comeycom_wfm:pricetravel2015@tcp(cundbwf01.pricetravel.com.mx:3306)/comeycom_WFM",
	"ccexporter.usr:IFaoCJiH09rEqLVZVLsj@tcp(cundbwf01.pricetravel.com.mx:3306)/ccexporter",
}
var okPaises = []string{
	"MX",
	"CO",
}

var functions = []string{
	"liveCalls",
	"tdDetails",
	"tdPauses",
	"test",
}

var fDb, fDbXp *sql.DB
var err error
var sp *spinner.Spinner

func main() {

	// Validacion de inputs
	if len(os.Args) < 2 {
		printInstructions("Especifica el tipo de proceso a correr:")
		listProcesos()
		return
	}

	proceso := os.Args[1]
	r, _ := inArray(proceso, functions)
	if !r {
		printInstructions(fmt.Sprintf("El proceso \"%s\" no existe, especifica uno de los siguientes:", proceso))
		listProcesos()
		return
	}

	sp = spinner.New(spinner.CharSets[43], 100*time.Millisecond) // Build our new spinner
	sp.Color("fgHiGreen")
	// Conexión a bases de datos

	fDb, err = sql.Open("mysql", dbCon[0])
	if err != nil {
		fmt.Printf(err.Error())

	}
	fDb.SetMaxOpenConns(4)
	fDb.SetMaxIdleConns(0)
	fDb.SetConnMaxLifetime(time.Second * 10)
	fDbXp, err = sql.Open("mysql", dbCon[1])
	if err != nil {
		fmt.Printf(err.Error())

	}
	fDbXp.SetMaxOpenConns(4)
	fDbXp.SetMaxIdleConns(4)
	fDbXp.SetConnMaxLifetime(time.Second * 10)

	defer fDb.Close()
	defer fDbXp.Close()

	switch proceso {
	case "liveCalls":
		liveCalls()
	case "tdDetails":
		runBackProcess()
	case "tdPauses":
		getPauses()
	default:
		printInstructions(fmt.Sprintf("El proceso \"%s\" no se encuentra aún activo\n", proceso))
		return
	}

}

func liveCalls() {

	sp.Stop()
	// Validacion de inputs
	paises := valPaises()

	// Inicio de LiveCalls
	printInstructions("Proceso: Live Calls")

	fmt.Println()
	runAgentCalls(paises)
	fmt.Println()
	runLiveCalls(paises)
	fmt.Println()

	flag := prReload(1)
	if flag {
		liveCalls()
	}

}

func printSlice(s []string) {
	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
}

func inArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

func listProcesos() {
	fmt.Println()
	for _, v := range functions {
		fmt.Printf("* %s\n", v)
	}
	fmt.Println()
}

func printInstructions(t string) {
	fmt.Println()
	fmt.Println("=========================================")
	fmt.Println()
	fmt.Printf("%s\n", t)
}

// Readln reads and returns a single line (sentinal: \n) from stdin.
// If a given timeout has passed, an error is returned.
// This functionality is similar to GNU's `read -e -p [s] -t [num] [name]`
func exitTO(prompt string, timeout time.Duration) (bool, error) {
	s := make(chan string)
	e := make(chan error)

	go func() {
		fmt.Print(prompt)
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			e <- err
		} else {
			s <- line
		}

		fmt.Printf("%s", line)
		result := line[0 : len(line)-1]
		if result == "x" {
			printInstructions("Programa finalizado\n\n")
			os.Exit(1)
		}

		close(s)
		close(e)
	}()

	select {
	case line := <-s:
		fmt.Printf("%s", line)
		result := line[0 : len(line)-1]
		if result == "x" {
			return true, nil
		}

		return false, nil
	case err := <-e:
		return false, err
	case <-time.After(timeout):
		return false, errors.New("Loop start")
	}
}

func exitPrompt(prompt string, seconds int) bool {
	m, _ := time.ParseDuration(fmt.Sprintf("%ds", seconds))

	flag, err := exitTO(prompt, m)

	if err != nil {
		fmt.Println(err.Error())
	}

	return flag
}

func prReload(t int) bool {
	fmt.Println()

	count := t

	// Original Bar
	// tmpl := fmt.Sprintf(`{{ red "Reload in %d seconds:" }} {{bar . | green}} {{counters . | blue }}`, t)
	// bar := pb.ProgressBarTemplate(tmpl).Start(count)
	// bar.SetWidth(80)
	// defer bar.Finish()

	// inc := count / t
	fmt.Printf("Wait %d seconds -> ", count)
	for i := 0; i < t; i++ {
		// bar.Add(1)
		fmt.Printf("%s", "#")
		time.Sleep(time.Second)
	}

	fmt.Println()
	fmt.Println()
	return true
}

func formatPrint(t string, f bool) {
	v := " "
	if f {
		v = "-"
	}
	d := fmt.Sprintf("|%s %s", v, t)
	l := 54 - utf8.RuneCountInString(d)
	fmt.Printf("%s", d)
	for i := 0; i < l; i++ {
		fmt.Printf("%s", " ")
	}
	fmt.Printf("%s\n", "|")

}

func printFrame(tl string, tp bool) {

	var fr, tpS string

	if tp {
		tpS = "START"
	} else {
		tpS = "END"
	}

	if utf8.RuneCountInString(tl)%2 != 0 {
		tl += " "
	}
	tl = fmt.Sprintf("%s %s", tl, tpS)
	tlLen := (54 - (utf8.RuneCountInString(tl) + 2)) / 2

	for i := 1; i <= tlLen; i++ {
		fr += "="
	}

	fmt.Printf("%[1]s %[2]s %[1]s\n", fr, tl)
}

func wrapTxt(t string, ln int) (m string) {
	var fr string
	wrapper := wordwrap.Wrapper(ln, false)
	t = wrapper(t)

	re := regexp.MustCompile(`\r?\n`)
	m = re.ReplaceAllString(t, "\n|            ")

	for i := 1; i <= 53-len(m); i++ {
		fr += " "
	}

	return
}

func printStatus(m string, err error) {
	if err != nil {
		formatPrint(wrapTxt(fmt.Sprintf("       ERROR!: %s", err.Error()), 40), false)
		formatPrint("", false)
	} else {
		formatPrint(fmt.Sprintf("       OK!"), false)
		if m != "" {
			formatPrint(fmt.Sprintf("       %s", m), false)
		}
		formatPrint("", false)
	}
}

func getFromQm(rqT string, uri string, block string, prefix int, arrFlag bool, data url.Values) (body []byte, fields []string, values [][]string, err error) {

	//Get data from QM (API)

	req, err := http.NewRequest(rqT, uri, bytes.NewBufferString(data.Encode()))
	req.Header.Set("content-type", `application/x-www-form-urlencoded; param=value`)
	req.Header.Add("Authorization", `Basic cm9ib3Q6cm9ib3Q=`)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	//Convert response to readable array
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if body == nil {
		err = fmt.Errorf("Sin resultados")
		return
	}

	// If array requested
	if arrFlag {

		var posts map[string]interface{}
		json.Unmarshal(body, &posts)

		if posts[block] == nil {
			err = fmt.Errorf("Sin bloques para construir")
			return
		}

		// Get Fields
		for _, v := range posts[block].([]interface{})[0].([]interface{}) {
			fields = append(fields, v.(string)[prefix:])
		}

		// Get Values
		for _, v := range posts[block].([]interface{})[1:] {
			var valor []string
			for _, r := range v.([]interface{}) {
				valor = append(valor, r.(string))
			}
			values = append(values, valor)
		}

	}

	return
}

func valPaises() (paises []string) {

	if len(os.Args) < 3 {
		printInstructions("Ingresa los paises que deseas obtener, separados por un espacio:\n")
		for _, v := range okPaises {
			fmt.Printf("* %s\n", v)
		}
		fmt.Println()
		os.Exit(2)
	}

	var omited []string
	for _, v := range os.Args[2:] {
		r, _ := inArray(v, okPaises)
		if r {
			paises = append(paises, v)
		} else {
			omited = append(omited, v)
		}
	}

	if len(paises) == 0 {
		printInstructions("No has ingresado paises válidos. Los paises permitidos son:\n")
		for i, v := range okPaises {
			fmt.Printf("%2d: %s\n", i, v)
		}
		fmt.Println()
		os.Exit(2)
	}

	if len(omited) > 0 {
		printInstructions("Paises omitidos por no estar en parámetros:\n")
		for _, v := range omited {
			fmt.Printf("%v | ", v)
		}
		fmt.Println()
	}

	return

}
