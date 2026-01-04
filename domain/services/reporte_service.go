package services

import (
	"bytes"
	"context"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/entities"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type ReporteService struct {
	clienteRepo     ports.PortCliente
	empresaRepo     ports.PortEmpresa
	dispositivoRepo ports.PortDispositivo
	boletaRepo      ports.PortBoleta
}

func NewReporteService(
	clienteRepo ports.PortCliente,
	empresaRepo ports.PortEmpresa,
	dispositivoRepo ports.PortDispositivo,
	boletaRepo ports.PortBoleta,
) *ReporteService {
	return &ReporteService{
		clienteRepo:     clienteRepo,
		empresaRepo:     empresaRepo,
		dispositivoRepo: dispositivoRepo,
		boletaRepo:      boletaRepo,
	}
}

func (s *ReporteService) GenerarReporteClientes(ctx context.Context, empresaID string, formato string) (*bytes.Buffer, string, error) {
	clientes, err := s.clienteRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, "", err
	}

	switch formato {
	case "excel":
		return s.generarExcelClientes(clientes)
	case "pdf":
		return s.generarPDFClientes(clientes)
	default:
		return s.generarCSVClientes(clientes)
	}
}

func (s *ReporteService) GenerarReporteDispositivos(ctx context.Context, empresaID string, formato string) (*bytes.Buffer, string, error) {
	dispositivosEntity, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, "", err
	}

	switch formato {
	case "excel":
		return s.generarExcelDispositivos(dispositivosEntity)
	case "pdf":
		return s.generarPDFDispositivos(dispositivosEntity)
	default:
		return s.generarCSVDispositivos(dispositivosEntity)
	}
}

func (s *ReporteService) GenerarReporteBoletas(ctx context.Context, clienteID string, formato string) (*bytes.Buffer, string, error) {
	boletasEntity, err := s.boletaRepo.FindByCliente(ctx, clienteID)
	if err != nil {
		return nil, "", err
	}

	switch formato {
	case "excel":
		return s.generarExcelBoletas(boletasEntity)
	case "pdf":
		return s.generarPDFBoletas(boletasEntity)
	default:
		return s.generarCSVBoletas(boletasEntity)
	}
}

func (s *ReporteService) GenerarReporteConsumo(ctx context.Context, empresaID string, fechaInicio, fechaFin time.Time, formato string) (*bytes.Buffer, string, error) {
	dispositivosEntity, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, "", err
	}

	switch formato {
	case "excel":
		return s.generarExcelConsumo(dispositivosEntity, fechaInicio, fechaFin)
	case "pdf":
		return s.generarPDFConsumo(dispositivosEntity, fechaInicio, fechaFin)
	default:
		return s.generarCSVConsumo(dispositivosEntity, fechaInicio, fechaFin)
	}
}

func (s *ReporteService) generarExcelClientes(clientes []*models.ClienteModel) (*bytes.Buffer, string, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Clientes"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	headers := []string{"Número Cliente", "Nombre", "Correo", "Teléfono", "Dirección", "Ciudad", "RUT", "Tipo", "Estado", "Fecha Creación"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	for i, cliente := range clientes {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), cliente.NumeroCliente)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), cliente.Nombre)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), cliente.Correo)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), cliente.Telefono)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), cliente.Direccion)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), cliente.Ciudad)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), cliente.Rut)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), cliente.TipoCliente)
		estado := "Inactivo"
		if cliente.Activo {
			estado = "Activo"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), estado)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), time.Now().Format("02/01/2006"))
	}

	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("reporte_clientes_%s.xlsx", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}

func (s *ReporteService) generarCSVClientes(clientes []*models.ClienteModel) (*bytes.Buffer, string, error) {
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	headers := []string{"Número Cliente", "Nombre", "Correo", "Teléfono", "Dirección", "Ciudad", "RUT", "Tipo", "Estado", "Fecha Creación"}
	writer.Write(headers)

	for _, cliente := range clientes {
		estado := "Inactivo"
		if cliente.Activo {
			estado = "Activo"
		}
		record := []string{
			cliente.NumeroCliente,
			cliente.Nombre,
			cliente.Correo,
			cliente.Telefono,
			cliente.Direccion,
			cliente.Ciudad,
			cliente.Rut,
			cliente.TipoCliente,
			estado,
			time.Now().Format("02/01/2006"),
		}
		writer.Write(record)
	}

	writer.Flush()
	filename := fmt.Sprintf("reporte_clientes_%s.csv", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}

func (s *ReporteService) generarExcelDispositivos(dispositivos []*entities.DispositivoEntity) (*bytes.Buffer, string, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Dispositivos"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	headers := []string{"Número Dispositivo", "Nombre", "Tipo", "Estado", "Consumo Actual (kWh)", "Voltaje (V)", "Corriente (A)", "Potencia Activa (W)"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	for i, dispositivo := range dispositivos {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), dispositivo.NumeroDispositivo)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), dispositivo.Nombre)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), dispositivo.Tipo)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), dispositivo.Estado)
		
		if dispositivo.UltimaLectura != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), dispositivo.UltimaLectura.Energy)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), dispositivo.UltimaLectura.Voltage)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), dispositivo.UltimaLectura.Current)
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), dispositivo.UltimaLectura.ActivePower)
		}
	}

	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("reporte_dispositivos_%s.xlsx", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}

func (s *ReporteService) generarCSVDispositivos(dispositivos []*entities.DispositivoEntity) (*bytes.Buffer, string, error) {
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	headers := []string{"Número Dispositivo", "Nombre", "Tipo", "Estado", "Consumo Actual (kWh)", "Voltaje (V)", "Corriente (A)", "Potencia Activa (W)"}
	writer.Write(headers)

	for _, dispositivo := range dispositivos {
		energy := ""
		voltage := ""
		current := ""
		power := ""
		
		if dispositivo.UltimaLectura != nil {
			energy = fmt.Sprintf("%.2f", dispositivo.UltimaLectura.Energy)
			voltage = fmt.Sprintf("%.2f", dispositivo.UltimaLectura.Voltage)
			current = fmt.Sprintf("%.2f", dispositivo.UltimaLectura.Current)
			power = fmt.Sprintf("%.2f", dispositivo.UltimaLectura.ActivePower)
		}

		record := []string{
			dispositivo.NumeroDispositivo,
			dispositivo.Nombre,
			dispositivo.Tipo,
			dispositivo.Estado,
			energy,
			voltage,
			current,
			power,
		}
		writer.Write(record)
	}

	writer.Flush()
	filename := fmt.Sprintf("reporte_dispositivos_%s.csv", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}

func (s *ReporteService) generarExcelBoletas(boletas []*entities.BoletaEntity) (*bytes.Buffer, string, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Boletas"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	headers := []string{"ID Boleta", "Cliente ID", "Período", "Monto Total", "Estado", "Fecha Creación", "Fecha Pago"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	for i, boleta := range boletas {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), boleta.ID.Hex())
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), boleta.ClienteID.Hex())
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), boleta.Periodo)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), boleta.Monto)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), boleta.Estado)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), boleta.FechaCreacion.Format("02/01/2006"))
		if boleta.FechaPago != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), boleta.FechaPago.Format("02/01/2006"))
		}
	}

	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("reporte_boletas_%s.xlsx", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}

func (s *ReporteService) generarCSVBoletas(boletas []*entities.BoletaEntity) (*bytes.Buffer, string, error) {
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	headers := []string{"ID Boleta", "Cliente ID", "Período", "Monto Total", "Estado", "Fecha Creación", "Fecha Pago"}
	writer.Write(headers)

	for _, boleta := range boletas {
		fechaPago := ""
		if boleta.FechaPago != nil {
			fechaPago = boleta.FechaPago.Format("02/01/2006")
		}
		
		record := []string{
			boleta.ID.Hex(),
			boleta.ClienteID.Hex(),
			boleta.Periodo,
			fmt.Sprintf("%.2f", boleta.Monto),
			boleta.Estado,
			boleta.FechaCreacion.Format("02/01/2006"),
			fechaPago,
		}
		writer.Write(record)
	}

	writer.Flush()
	filename := fmt.Sprintf("reporte_boletas_%s.csv", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}

func (s *ReporteService) generarExcelConsumo(dispositivos []*entities.DispositivoEntity, fechaInicio, fechaFin time.Time) (*bytes.Buffer, string, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Consumo"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	headers := []string{"Dispositivo", "Tipo", "Estado", "Consumo Total (kWh)", "Costo", "Período"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	periodo := fmt.Sprintf("%s - %s", fechaInicio.Format("02/01/2006"), fechaFin.Format("02/01/2006"))

	for i, dispositivo := range dispositivos {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), dispositivo.Nombre)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), dispositivo.Tipo)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), dispositivo.Estado)
		
		if dispositivo.UltimaLectura != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), dispositivo.UltimaLectura.Energy)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), dispositivo.UltimaLectura.Cost)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), periodo)
	}

	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("reporte_consumo_%s.xlsx", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}

func (s *ReporteService) generarCSVConsumo(dispositivos []*entities.DispositivoEntity, fechaInicio, fechaFin time.Time) (*bytes.Buffer, string, error) {
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	headers := []string{"Dispositivo", "Tipo", "Estado", "Consumo Total (kWh)", "Costo", "Período"}
	writer.Write(headers)

	periodo := fmt.Sprintf("%s - %s", fechaInicio.Format("02/01/2006"), fechaFin.Format("02/01/2006"))

	for _, dispositivo := range dispositivos {
		energy := ""
		cost := ""
		if dispositivo.UltimaLectura != nil {
			energy = fmt.Sprintf("%.2f", dispositivo.UltimaLectura.Energy)
			cost = fmt.Sprintf("%.2f", dispositivo.UltimaLectura.Cost)
		}

		record := []string{
			dispositivo.Nombre,
			dispositivo.Tipo,
			dispositivo.Estado,
			energy,
			cost,
			periodo,
		}
		writer.Write(record)
	}

	writer.Flush()
	filename := fmt.Sprintf("reporte_consumo_%s.csv", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}


func (s *ReporteService) generarPDFClientes(clientes []*models.ClienteModel) (*bytes.Buffer, string, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Reporte de Clientes")
	pdf.Ln(15)
	
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(30, 7, "N° Cliente", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 7, "Nombre", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 7, "Correo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Telefono", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Ciudad", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "RUT", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Estado", "1", 1, "C", true, 0, "")
	
	pdf.SetFont("Arial", "", 9)
	for _, cliente := range clientes {
		estado := "Inactivo"
		if cliente.Activo {
			estado = "Activo"
		}
		pdf.CellFormat(30, 6, cliente.NumeroCliente, "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 6, cliente.Nombre, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 6, cliente.Correo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, cliente.Telefono, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, cliente.Ciudad, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, cliente.Rut, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, estado, "1", 1, "C", false, 0, "")
	}
	
	buf := new(bytes.Buffer)
	err := pdf.Output(buf)
	if err != nil {
		return nil, "", err
	}
	
	filename := fmt.Sprintf("reporte_clientes_%s.pdf", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}

func (s *ReporteService) generarPDFDispositivos(dispositivos []*entities.DispositivoEntity) (*bytes.Buffer, string, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Reporte de Dispositivos")
	pdf.Ln(15)
	
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(35, 7, "N° Dispositivo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 7, "Nombre", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Tipo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Estado", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Consumo (kWh)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Voltaje (V)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Corriente (A)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Potencia (W)", "1", 1, "C", true, 0, "")
	
	pdf.SetFont("Arial", "", 9)
	for _, dispositivo := range dispositivos {
		pdf.CellFormat(35, 6, dispositivo.NumeroDispositivo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 6, dispositivo.Nombre, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, dispositivo.Tipo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, dispositivo.Estado, "1", 0, "C", false, 0, "")
		
		if dispositivo.UltimaLectura != nil {
			pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", dispositivo.UltimaLectura.Energy), "1", 0, "R", false, 0, "")
			pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", dispositivo.UltimaLectura.Voltage), "1", 0, "R", false, 0, "")
			pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", dispositivo.UltimaLectura.Current), "1", 0, "R", false, 0, "")
			pdf.CellFormat(40, 6, fmt.Sprintf("%.2f", dispositivo.UltimaLectura.ActivePower), "1", 1, "R", false, 0, "")
		} else {
			pdf.CellFormat(30, 6, "-", "1", 0, "C", false, 0, "")
			pdf.CellFormat(30, 6, "-", "1", 0, "C", false, 0, "")
			pdf.CellFormat(30, 6, "-", "1", 0, "C", false, 0, "")
			pdf.CellFormat(40, 6, "-", "1", 1, "C", false, 0, "")
		}
	}
	
	buf := new(bytes.Buffer)
	err := pdf.Output(buf)
	if err != nil {
		return nil, "", err
	}
	
	filename := fmt.Sprintf("reporte_dispositivos_%s.pdf", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}

func (s *ReporteService) generarPDFBoletas(boletas []*entities.BoletaEntity) (*bytes.Buffer, string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Reporte de Boletas")
	pdf.Ln(15)
	
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(40, 7, "ID Boleta", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Periodo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Monto", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Estado", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Fecha", "1", 1, "C", true, 0, "")
	
	pdf.SetFont("Arial", "", 9)
	totalMonto := 0.0
	for _, boleta := range boletas {
		totalMonto += boleta.Monto
		pdf.CellFormat(40, 6, boleta.ID.Hex()[:12]+"...", "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, boleta.Periodo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, fmt.Sprintf("$%.2f", boleta.Monto), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 6, boleta.Estado, "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 6, boleta.FechaCreacion.Format("02/01/2006"), "1", 1, "C", false, 0, "")
	}
	
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(110, 7, "TOTAL:")
	pdf.Cell(30, 7, fmt.Sprintf("$%.2f", totalMonto))
	
	buf := new(bytes.Buffer)
	err := pdf.Output(buf)
	if err != nil {
		return nil, "", err
	}
	
	filename := fmt.Sprintf("reporte_boletas_%s.pdf", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}

func (s *ReporteService) generarPDFConsumo(dispositivos []*entities.DispositivoEntity, fechaInicio, fechaFin time.Time) (*bytes.Buffer, string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Reporte de Consumo Electrico")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 11)
	periodo := fmt.Sprintf("Periodo: %s - %s", fechaInicio.Format("02/01/2006"), fechaFin.Format("02/01/2006"))
	pdf.Cell(0, 8, periodo)
	pdf.Ln(12)
	
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(60, 7, "Dispositivo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Tipo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Estado", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Consumo (kWh)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Costo ($)", "1", 1, "C", true, 0, "")
	
	pdf.SetFont("Arial", "", 9)
	totalConsumo := 0.0
	totalCosto := 0.0
	
	for _, dispositivo := range dispositivos {
		pdf.CellFormat(60, 6, dispositivo.Nombre, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, dispositivo.Tipo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, dispositivo.Estado, "1", 0, "C", false, 0, "")
		
		if dispositivo.UltimaLectura != nil {
			totalConsumo += dispositivo.UltimaLectura.Energy
			totalCosto += dispositivo.UltimaLectura.Cost
			pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", dispositivo.UltimaLectura.Energy), "1", 0, "R", false, 0, "")
			pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", dispositivo.UltimaLectura.Cost), "1", 1, "R", false, 0, "")
		} else {
			pdf.CellFormat(30, 6, "-", "1", 0, "C", false, 0, "")
			pdf.CellFormat(30, 6, "-", "1", 1, "C", false, 0, "")
		}
	}
	
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(130, 7, "TOTAL CONSUMO:")
	pdf.Cell(30, 7, fmt.Sprintf("%.2f kWh", totalConsumo))
	pdf.Ln(8)
	pdf.Cell(130, 7, "TOTAL COSTO:")
	pdf.Cell(30, 7, fmt.Sprintf("$%.2f", totalCosto))
	
	buf := new(bytes.Buffer)
	err := pdf.Output(buf)
	if err != nil {
		return nil, "", err
	}
	
	filename := fmt.Sprintf("reporte_consumo_%s.pdf", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}
