package export

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

func (s *PDFService) ExportarClientes(data []map[string]interface{}) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Reporte de Clientes")
	pdf.Ln(12)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 6, fmt.Sprintf("Fecha: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(50, 7, "Nombre", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 7, "Correo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Telefono", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Tipo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Estado", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 8)
	for _, row := range data {
		estado := "Inactivo"
		if activo, ok := row["activo"].(bool); ok && activo {
			estado = "Activo"
		}
		
		pdf.CellFormat(50, 6, fmt.Sprintf("%v", row["nombre"]), "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 6, fmt.Sprintf("%v", row["correo"]), "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%v", row["telefono"]), "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%v", row["tipoCliente"]), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, estado, "1", 1, "C", false, 0, "")
	}

	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 5, fmt.Sprintf("Total de clientes: %d", len(data)))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *PDFService) ExportarDispositivos(data []map[string]interface{}) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(277, 10, "Reporte de Dispositivos")
	pdf.Ln(12)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(277, 6, fmt.Sprintf("Fecha: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(50, 7, "ID Dispositivo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Tipo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Modelo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 7, "Cliente", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Estado", "1", 0, "C", true, 0, "")
	pdf.CellFormat(32, 7, "Consumo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Instalacion", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 8)
	for _, row := range data {
		pdf.CellFormat(50, 6, fmt.Sprintf("%v", row["idDispositivo"]), "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("%v", row["tipo"]), "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%v", row["modelo"]), "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 6, fmt.Sprintf("%v", row["cliente"]), "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, fmt.Sprintf("%v", row["estado"]), "1", 0, "C", false, 0, "")
		pdf.CellFormat(32, 6, fmt.Sprintf("%v kWh", row["consumoActual"]), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 6, fmt.Sprintf("%v", row["fechaInstalacion"]), "1", 1, "C", false, 0, "")
	}

	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(277, 5, fmt.Sprintf("Total de dispositivos: %d", len(data)))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *PDFService) ExportarBoleta(data map[string]interface{}) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(190, 15, "BOLETA DE CONSUMO ELECTRICO")
	pdf.Ln(20)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, fmt.Sprintf("Boleta N: %v", data["numero"]))
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 6, fmt.Sprintf("Fecha Emision: %v", data["fechaEmision"]))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Fecha Vencimiento: %v", data["fechaVencimiento"]))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(190, 8, "Datos del Cliente")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 6, fmt.Sprintf("Nombre: %v", data["cliente"]))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Direccion: %v", data["direccion"]))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(190, 8, "Detalle de Consumo")
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(95, 7, "Concepto", "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 7, "Valor", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(95, 6, "Periodo", "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("%v", data["periodo"]), "1", 1, "R", false, 0, "")
	
	pdf.CellFormat(95, 6, "Consumo Total (kWh)", "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("%.2f kWh", data["consumo"]), "1", 1, "R", false, 0, "")
	
	pdf.CellFormat(95, 6, "Tarifa por kWh", "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("$%.2f", data["tarifa"]), "1", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(95, 8, "TOTAL A PAGAR", "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("$%.2f", data["monto"]), "1", 1, "R", true, 0, "")

	pdf.Ln(15)
	pdf.SetFont("Arial", "I", 8)
	pdf.MultiCell(190, 5, "Esta boleta es un documento tributario. Conserve este documento para cualquier reclamo o consulta.", "", "L", false)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
