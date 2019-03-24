package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/cheggaaa/pb.v2"
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
	"test",
}

var fDb, fDbXp *sql.DB
var err error

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

	// Conexión a bases de datos

	fDb, err = sql.Open("mysql", dbCon[0])
	if err != nil {
		panic(err.Error())
	}
	fDb.SetMaxOpenConns(4)
	fDb.SetMaxIdleConns(0)
	fDb.SetConnMaxLifetime(time.Second * 10)
	fDbXp, err = sql.Open("mysql", dbCon[1])
	if err != nil {
		panic(err.Error())
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
	default:
		printInstructions(fmt.Sprintf("El proceso \"%s\" no se encuentra aún activo\n", proceso))
		return
	}

}

func liveCalls() {

	// Validacion de inputs
	if len(os.Args) < 3 {
		printInstructions("Ingresa los paises que deseas obtener, separados por un espacio:\n")
		for _, v := range okPaises {
			fmt.Printf("* %s\n", v)
		}
		fmt.Println()
		return
	}

	var paises []string
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
		return
	}

	if len(omited) > 0 {
		printInstructions("Paises omitidos por no estar en parámetros:\n")
		for _, v := range omited {
			fmt.Printf("%v | ", v)
		}
		fmt.Println()
	}

	// Inicio de LiveCalls
	printInstructions("Proceso: Live Calls")

	fmt.Println()
	runAgentCalls(paises)
	fmt.Println()
	runLiveCalls(paises)
	fmt.Println()

	if exitPrompt("Reinicia en 1 segundo...", 1) {
		printInstructions("Programa finalizado\n\n")
		return
	}

	liveCalls()
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
	count := t
	tmpl := fmt.Sprintf(`{{ red "Reload in %d seconds:" }} {{bar . | green}} {{counters . | blue }}`, t)
	bar := pb.ProgressBarTemplate(tmpl).Start(count)
	bar.SetWidth(80)
	defer bar.Finish()

	// inc := count / t
	for i := 0; i < t; i++ {
		bar.Add(1)
		time.Sleep(time.Second)
	}
	return true
}
