package main

import (
	"fmt"
	"net/url"
)

const apiTst = "http://testoperaciones.pricetravel.com.mx/api/restfulbck/index.php/"
const apiProd = "https://operaciones.pricetravel.com.mx/api/restful/index.php/"

func runBackProcess() {

	// Inicio de LiveCalls
	printInstructions("Proceso: Back Proccesses")

	fmt.Println()
	mailing()
	fmt.Println()
	updateSchedules()
	fmt.Println()
	updateHX()
	fmt.Println()

	flag := prReload(30)
	if flag {
		runBackProcess()
	}
}

func mailing() {

	fmt.Printf("================== Mail lists START ==================\n")
	formatPrint("", false)

	params := url.Values{}

	// Contratos
	getDetail("Contratos:", "GET", "Mailing/contratosVencidos", apiTst, params)

	// Faltas Consecutivas
	getDetail("Faltas Consecutivas:", "GET", "Mailing/faltasConsecutivas", apiTst, params)

	// Revision de cumpleaneros Hoy
	getDetail("Revision de cumpleañeros Hoy:", "GET", "Mailing/cumpleHoy", apiTst, params)

	// Revision de cumpleaneros Personalizado
	getDetail("Revision de cumpleañeros Personalizado:", "GET", "Mailing/cumplePersonalizado", apiTst, params)

	// Revision de cumpleaneros Mes
	getDetail("Revision de cumpleañeros Mes:", "GET", "Mailing/cumpleMes", apiTst, params)

	fmt.Println("|")
	fmt.Printf("================== Mail lists   END ==================\n")

}

func updateSchedules() {

	fmt.Printf("=============== Update Schedules START ===============\n")
	formatPrint("", false)
	formatPrint("Corriendo Query:", true)

	u := "SELECT TIMETODATETIME(id) FROM `Historial Programacion` WHERE LastUpdate >= CAST(CONCAT(CURDATE(),' 00:00:00') as DATETIME)"
	r, err := fDb.Exec(u)
	if err != nil {
		formatPrint(fmt.Sprintf("       ERROR!: %s", err.Error()), false)
		formatPrint("", false)
	} else {
		formatPrint(fmt.Sprintf("       OK!"), false)
		formatPrint("", false)
	}
	_ = r

	formatPrint("", false)
	fmt.Printf("=============== Update Schedules   END ===============\n")
}

func updateHX() {

	fmt.Printf("================== Update HX  START ==================\n")
	formatPrint("", false)
	formatPrint("Corriendo Query:", true)
	formatPrint("", false)

	u := `UPDATE asesores_programacion 
    SET  
        phx_done = IF(COALESCE(TIME_TO_SEC(TIMEDIFF(IF(CHECKLOG(Fecha, asesor, 'in') < x1e 
                                            AND CHECKLOG(Fecha, asesor, 'out') > x1s, 
                                        IF(CHECKLOG(Fecha, asesor, 'out') > x1e, 
                                            x1e, 
                                            CHECKLOG(Fecha, asesor, 'out')), 
                                        NULL), 
                                    IF(CHECKLOG(Fecha, asesor, 'in') < x1e 
                                            AND CHECKLOG(Fecha, asesor, 'out') > x1s, 
                                        IF(CHECKLOG(Fecha, asesor, 'in') < x1s, 
                                            x1s, 
                                            CHECKLOG(Fecha, asesor, 'in')), 
                                        NULL))) / 60 / 60, 
                    0) + COALESCE(TIME_TO_SEC(TIMEDIFF(IF(CHECKLOG(Fecha, asesor, 'in') < x2e 
                                            AND CHECKLOG(Fecha, asesor, 'out') > x2s, 
                                        IF(CHECKLOG(Fecha, asesor, 'out') > x2e, 
                                            x2e, 
                                            CHECKLOG(Fecha, asesor, 'out')), 
                                        NULL), 
                                    IF(CHECKLOG(Fecha, asesor, 'in') < x2e 
                                            AND CHECKLOG(Fecha, asesor, 'out') > x2s, 
                                        IF(CHECKLOG(Fecha, asesor, 'in') < x2s, 
                                            x2s, 
                                            CHECKLOG(Fecha, asesor, 'in')), 
                                        NULL))) / 60 / 60, 
                    0) > COALESCE(TIMEDIFF(x1e, x1s), 0) + COALESCE(TIMEDIFF(x2e, x2s), 0), 
            COALESCE(TIMEDIFF(x1e, x1s), 0) + COALESCE(TIMEDIFF(x2e, x2s), 0), 
            COALESCE(TIME_TO_SEC(TIMEDIFF(IF(CHECKLOG(Fecha, asesor, 'in') < x1e 
                                            AND CHECKLOG(Fecha, asesor, 'out') > x1s, 
                                        IF(CHECKLOG(Fecha, asesor, 'out') > x1e, 
                                            x1e, 
                                            CHECKLOG(Fecha, asesor, 'out')), 
                                        NULL), 
                                    IF(CHECKLOG(Fecha, asesor, 'in') < x1e 
                                            AND CHECKLOG(Fecha, asesor, 'out') > x1s, 
                                        IF(CHECKLOG(Fecha, asesor, 'in') < x1s, 
                                            x1s, 
                                            CHECKLOG(Fecha, asesor, 'in')), 
                                        NULL))) / 60 / 60, 
                    0) + COALESCE(TIME_TO_SEC(TIMEDIFF(IF(CHECKLOG(Fecha, asesor, 'in') < x2e 
                                            AND CHECKLOG(Fecha, asesor, 'out') > x2s, 
                                        IF(CHECKLOG(Fecha, asesor, 'out') > x2e, 
                                            x2e, 
                                            CHECKLOG(Fecha, asesor, 'out')), 
                                        NULL), 
                                    IF(CHECKLOG(Fecha, asesor, 'in') < x2e 
                                            AND CHECKLOG(Fecha, asesor, 'out') > x2s, 
                                        IF(CHECKLOG(Fecha, asesor, 'in') < x2s, 
                                            x2s, 
                                            CHECKLOG(Fecha, asesor, 'in')), 
                                        NULL))) / 60 / 60, 
                    0)) 
    WHERE 
        (x1s != x1e OR x2s != x2e) 
            AND x1s >= ADDDATE(CURDATE(), - 20) `
	r, err := fDb.Exec(u)
	if err != nil {
		formatPrint(fmt.Sprintf("       ERROR!: %s", err.Error()), false)
		formatPrint("", false)
	} else {
		formatPrint(fmt.Sprintf("       OK!"), false)
		formatPrint("", false)
	}
	_ = r

	formatPrint("", false)
	fmt.Printf("================== Update HX    END ==================\n")
}

func getDetail(n string, tp string, route string, uri string, params url.Values) {
	// Contratos
	block := n
	formatPrint(block, true)

	m, err := getAPI(tp, route, uri, params)
	if err != nil {
		formatPrint(fmt.Sprintf("       ERROR!: %s", err.Error()), false)
		formatPrint("", false)
	} else {
		formatPrint(fmt.Sprintf("       OK!"), false)
		formatPrint(fmt.Sprintf("        %s", m), false)
		formatPrint("", false)
	}
}

func formatPrint(t string, f bool) {
	v := " "
	if f {
		v = "-"
	}
	d := fmt.Sprintf("|%s %s", v, t)
	l := 54 - len(d)
	fmt.Printf("%s", d)
	for i := 0; i < l; i++ {
		fmt.Printf("%s", " ")
	}
	fmt.Printf("%s\n", "|")

}
