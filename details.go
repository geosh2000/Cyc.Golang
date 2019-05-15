package main

import (
	"fmt"
	"net/url"
)

const apiTst = "http://testoperaciones.pricetravel.com.mx/api/restfulbck/index.php/"
const apiProd = "https://operaciones.pricetravel.com.mx/api/restful/index.php/"

const dtDone = "UPDATE asesores_ausentismos a LEFT JOIN asesores_programacion b ON a.asesor = b.asesor AND a.Fecha = b.Fecha SET pdt_done = COALESCE(TIME_TO_SEC(TIMEDIFF(IF(CHECKLOG(a.Fecha, a.asesor, 'out') > je, je, CHECKLOG(a.Fecha, a.asesor, 'out')), IF(CHECKLOG(a.Fecha, a.asesor, 'in') < js, js, CHECKLOG(a.Fecha, a.asesor, 'in')))) / TIME_TO_SEC(TIMEDIFF(je, js)) * 8,0) WHERE ausentismo = 19 AND a = 1 AND a.Fecha >= ADDDATE(CURDATE(), - 60)"
const updtScheds = "SELECT TIMETODATETIME(id) FROM `Historial Programacion` WHERE LastUpdate >= CAST(CONCAT(CURDATE(),' 00:00:00') as DATETIME)"
const updtHX = `UPDATE asesores_programacion 
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

func runBackProcess() {
	sp.Stop()
	// Inicio de LiveCalls
	printInstructions("Proceso: Back Proccesses")

	fmt.Println()
	mailing()
	fmt.Println()
	queryExe("Update Schedules", updtScheds)
	fmt.Println()
	queryExe("Update HX", updtHX)
	fmt.Println()
	queryExe("Update DTs", dtDone)
	fmt.Println()
	runAgDetails()
	fmt.Println()

	flag := prReload(30)
	if flag {
		runBackProcess()
	}
}

func mailing() {

	printFrame("Mail List", true)
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

	formatPrint("", false)
	printFrame("Mail List", false)

}

func getDetail(n string, tp string, route string, uri string, params url.Values) {
	// Contratos
	block := n
	formatPrint(block, true)

	sp.Start()
	m, err := getAPI(tp, route, uri, params)
	sp.Stop()

	printStatus(m, err)
}

func queryExe(tl string, qr string) {

	printFrame(tl, true)
	formatPrint("", false)
	formatPrint("Corriendo Query:", true)

	sp.Start()
	r, err := fDb.Exec(qr)
	sp.Stop()

	printStatus("", err)
	_ = r

	formatPrint("", false)
	printFrame(tl, false)

}
