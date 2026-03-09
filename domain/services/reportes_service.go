package services

import (
	"bytes"
	"context"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/entities"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

// ReportesService — sistema unificado de reportes
// Reemplaza ExportService y ReporteService
type ReportesService struct {
	clienteRepo      ports.PortCliente
	dispositivoRepo  ports.PortDispositivo
	notificacionRepo ports.PortNotificacion
	boletaRepo       ports.PortBoleta
}

func NewReportesService(
	clienteRepo ports.PortCliente,
	dispositivoRepo ports.PortDispositivo,
	notificacionRepo ports.PortNotificacion,
	boletaRepo ports.PortBoleta,
) *ReportesService {
	return &ReportesService{
		clienteRepo:      clienteRepo,
		dispositivoRepo:  dispositivoRepo,
		notificacionRepo: notificacionRepo,
		boletaRepo:       boletaRepo,
	}
}

// ─── Clientes ─────────────────────────────────────────────────────────────────

func (s *ReportesService) ReporteClientes(ctx context.Context, empresaID, formato string) ([]byte, string, string, error) {
	clientes, err := s.clienteRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, "", "", err
	}
	switch formato {
	case "pdf":
		data, name, err := clientesPDF(clientes)
		return data, name, "application/pdf", err
	default:
		data, name, err := clientesExcel(clientes)
		return data, name, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", err
	}
}

func clientesExcel(clientes []*models.ClienteModel) ([]byte, string, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Clientes"
	idx, _ := f.NewSheet(sheet)
	f.SetActiveSheet(idx)

	headers := []string{"N° Cliente", "Nombre", "Correo", "Teléfono", "Dirección", "Ciudad", "RUT", "Tipo", "Estado", "Fecha Registro"}
	for i, h := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), h)
	}
	for i, c := range clientes {
		r := i + 2
		estado := "Inactivo"
		if c.Activo {
			estado = "Activo"
		}
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), c.NumeroCliente)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), c.Nombre)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), c.Correo)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), c.Telefono)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), c.Direccion)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", r), c.Ciudad)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", r), c.Rut)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", r), c.TipoCliente)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", r), estado)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", r), c.FechaRegistro.Format("02/01/2006"))
	}
	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), fmt.Sprintf("clientes_%s.xlsx", time.Now().Format("20060102_150405")), nil
}

func clientesPDF(clientes []*models.ClienteModel) ([]byte, string, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Reporte de Clientes")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(20, 20, 20)
	pdf.SetTextColor(255, 255, 255)
	cols := []struct {
		w float64
		t string
	}{
		{30, "N° Cliente"}, {50, "Nombre"}, {60, "Correo"},
		{28, "Teléfono"}, {38, "Ciudad"}, {25, "RUT"}, {25, "Estado"},
	}
	for _, c := range cols {
		pdf.CellFormat(c.w, 7, c.t, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(0, 0, 0)
	for _, c := range clientes {
		estado := "Inactivo"
		if c.Activo {
			estado = "Activo"
		}
		pdf.CellFormat(30, 6, c.NumeroCliente, "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 6, c.Nombre, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 6, c.Correo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(28, 6, c.Telefono, "1", 0, "L", false, 0, "")
		pdf.CellFormat(38, 6, c.Ciudad, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, c.Rut, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, estado, "1", 1, "C", false, 0, "")
	}
	buf := new(bytes.Buffer)
	if err := pdf.Output(buf); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), fmt.Sprintf("clientes_%s.pdf", time.Now().Format("20060102_150405")), nil
}

// ─── Dispositivos ─────────────────────────────────────────────────────────────

func (s *ReportesService) ReporteDispositivos(ctx context.Context, empresaID, formato string) ([]byte, string, string, error) {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, "", "", err
	}
	switch formato {
	case "pdf":
		data, name, err := dispositivosPDF(dispositivos)
		return data, name, "application/pdf", err
	default:
		data, name, err := dispositivosExcel(dispositivos)
		return data, name, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", err
	}
}

func dispositivosExcel(dispositivos []*entities.DispositivoEntity) ([]byte, string, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Dispositivos"
	idx, _ := f.NewSheet(sheet)
	f.SetActiveSheet(idx)

	headers := []string{"N° Dispositivo", "Nombre", "Tipo", "Estado", "Consumo (kWh)", "Voltaje (V)", "Corriente (A)", "Potencia (W)", "Costo (CLP)"}
	for i, h := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), h)
	}
	for i, d := range dispositivos {
		r := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), d.NumeroDispositivo)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), d.Nombre)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), d.Tipo)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), d.Estado)
		if d.UltimaLectura != nil {
			f.SetCellValue(sheet, fmt.Sprintf("E%d", r), d.UltimaLectura.Energy)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", r), d.UltimaLectura.Voltage)
			f.SetCellValue(sheet, fmt.Sprintf("G%d", r), d.UltimaLectura.Current)
			f.SetCellValue(sheet, fmt.Sprintf("H%d", r), d.UltimaLectura.ActivePower)
			f.SetCellValue(sheet, fmt.Sprintf("I%d", r), d.UltimaLectura.Cost)
		}
	}
	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), fmt.Sprintf("dispositivos_%s.xlsx", time.Now().Format("20060102_150405")), nil
}

func dispositivosPDF(dispositivos []*entities.DispositivoEntity) ([]byte, string, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Reporte de Dispositivos")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(20, 20, 20)
	pdf.SetTextColor(255, 255, 255)
	cols := []struct {
		w float64
		t string
	}{
		{35, "N° Disp."}, {45, "Nombre"}, {25, "Tipo"}, {25, "Estado"},
		{32, "Consumo kWh"}, {25, "Voltaje V"}, {25, "Corriente A"}, {28, "Potencia W"},
	}
	for _, c := range cols {
		pdf.CellFormat(c.w, 7, c.t, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(0, 0, 0)
	for _, d := range dispositivos {
		pdf.CellFormat(35, 6, d.NumeroDispositivo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(45, 6, d.Nombre, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, d.Tipo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, d.Estado, "1", 0, "C", false, 0, "")
		if d.UltimaLectura != nil {
			pdf.CellFormat(32, 6, fmt.Sprintf("%.3f", d.UltimaLectura.Energy), "1", 0, "R", false, 0, "")
			pdf.CellFormat(25, 6, fmt.Sprintf("%.1f", d.UltimaLectura.Voltage), "1", 0, "R", false, 0, "")
			pdf.CellFormat(25, 6, fmt.Sprintf("%.2f", d.UltimaLectura.Current), "1", 0, "R", false, 0, "")
			pdf.CellFormat(28, 6, fmt.Sprintf("%.1f", d.UltimaLectura.ActivePower), "1", 1, "R", false, 0, "")
		} else {
			pdf.CellFormat(32, 6, "-", "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 6, "-", "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 6, "-", "1", 0, "C", false, 0, "")
			pdf.CellFormat(28, 6, "-", "1", 1, "C", false, 0, "")
		}
	}
	buf := new(bytes.Buffer)
	if err := pdf.Output(buf); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), fmt.Sprintf("dispositivos_%s.pdf", time.Now().Format("20060102_150405")), nil
}

// ─── Alertas ──────────────────────────────────────────────────────────────────

func (s *ReportesService) ReporteAlertas(ctx context.Context, empresaID, formato string) ([]byte, string, string, error) {
	alertas, err := s.notificacionRepo.FindByEmpresa(ctx, empresaID)
	if err != nil {
		return nil, "", "", err
	}
	switch formato {
	case "pdf":
		data, name, err := alertasPDF(alertas)
		return data, name, "application/pdf", err
	default:
		data, name, err := alertasExcel(alertas)
		return data, name, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", err
	}
}

func alertasExcel(alertas []*entities.NotificacionEntity) ([]byte, string, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Alertas"
	idx, _ := f.NewSheet(sheet)
	f.SetActiveSheet(idx)

	headers := []string{"Tipo", "Título", "Mensaje", "Severidad", "Resuelta", "Fecha Creación", "Fecha Resolución"}
	for i, h := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), h)
	}
	for i, a := range alertas {
		r := i + 2
		resuelta := "No"
		if a.Resuelta {
			resuelta = "Sí"
		}
		fechaRes := ""
		if a.FechaResolucion != nil {
			fechaRes = a.FechaResolucion.Format("02/01/2006 15:04")
		}
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), a.Tipo)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), a.Titulo)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), a.Mensaje)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), a.Severidad)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), resuelta)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", r), a.FechaCreacion.Format("02/01/2006 15:04"))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", r), fechaRes)
	}
	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), fmt.Sprintf("alertas_%s.xlsx", time.Now().Format("20060102_150405")), nil
}

func alertasPDF(alertas []*entities.NotificacionEntity) ([]byte, string, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Reporte de Alertas")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(20, 20, 20)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(35, 7, "Tipo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(90, 7, "Mensaje", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Severidad", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Resuelta", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 7, "Fecha Creación", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 7, "Fecha Resolución", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(0, 0, 0)
	for _, a := range alertas {
		resuelta := "No"
		if a.Resuelta {
			resuelta = "Sí"
		}
		fechaRes := ""
		if a.FechaResolucion != nil {
			fechaRes = a.FechaResolucion.Format("02/01/2006 15:04")
		}
		pdf.CellFormat(35, 6, a.Tipo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(90, 6, a.Mensaje, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, a.Severidad, "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, resuelta, "1", 0, "C", false, 0, "")
		pdf.CellFormat(45, 6, a.FechaCreacion.Format("02/01/2006 15:04"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(45, 6, fechaRes, "1", 1, "C", false, 0, "")
	}
	buf := new(bytes.Buffer)
	if err := pdf.Output(buf); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), fmt.Sprintf("alertas_%s.pdf", time.Now().Format("20060102_150405")), nil
}

// ─── Boletas ──────────────────────────────────────────────────────────────────

func (s *ReportesService) ReporteBoletas(ctx context.Context, empresaID, formato string) ([]byte, string, string, error) {
	clientes, err := s.clienteRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, "", "", err
	}

	type row struct {
		cliente string
		periodo string
		monto   float64
		estado  string
		fecha   string
	}
	var rows []row
	for _, c := range clientes {
		boletas, err := s.boletaRepo.FindByCliente(ctx, c.ID)
		if err != nil {
			continue
		}
		for _, b := range boletas {
			rows = append(rows, row{
				cliente: c.Nombre,
				periodo: b.Periodo,
				monto:   b.Monto,
				estado:  b.Estado,
				fecha:   b.FechaCreacion.Format("02/01/2006"),
			})
		}
	}

	switch formato {
	case "pdf":
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Reporte de Boletas")
		pdf.Ln(12)

		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(20, 20, 20)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(55, 7, "Cliente", "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 7, "Período", "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 7, "Monto", "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 7, "Estado", "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 7, "Fecha", "1", 1, "C", true, 0, "")

		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(0, 0, 0)
		total := 0.0
		for _, r := range rows {
			total += r.monto
			pdf.CellFormat(55, 6, r.cliente, "1", 0, "L", false, 0, "")
			pdf.CellFormat(35, 6, r.periodo, "1", 0, "L", false, 0, "")
			pdf.CellFormat(35, 6, fmt.Sprintf("$%.0f", r.monto), "1", 0, "R", false, 0, "")
			pdf.CellFormat(30, 6, r.estado, "1", 0, "C", false, 0, "")
			pdf.CellFormat(35, 6, r.fecha, "1", 1, "C", false, 0, "")
		}
		pdf.Ln(4)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 7, fmt.Sprintf("TOTAL: $%.0f", total))

		buf := new(bytes.Buffer)
		if err := pdf.Output(buf); err != nil {
			return nil, "", "", err
		}
		return buf.Bytes(), fmt.Sprintf("boletas_%s.pdf", time.Now().Format("20060102_150405")), "application/pdf", nil

	default:
		f := excelize.NewFile()
		defer f.Close()
		sheet := "Boletas"
		idx, _ := f.NewSheet(sheet)
		f.SetActiveSheet(idx)

		headers := []string{"Cliente", "Período", "Monto", "Estado", "Fecha"}
		for i, h := range headers {
			f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), h)
		}
		for i, r := range rows {
			row := i + 2
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), r.cliente)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), r.periodo)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), r.monto)
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), r.estado)
			f.SetCellValue(sheet, fmt.Sprintf("E%d", row), r.fecha)
		}
		buf := new(bytes.Buffer)
		if err := f.Write(buf); err != nil {
			return nil, "", "", err
		}
		return buf.Bytes(), fmt.Sprintf("boletas_%s.xlsx", time.Now().Format("20060102_150405")),
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	}
}

// ─── Consumo ──────────────────────────────────────────────────────────────────

func (s *ReportesService) ReporteConsumo(ctx context.Context, empresaID string, fechaInicio, fechaFin time.Time, formato string) ([]byte, string, string, error) {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, "", "", err
	}
	periodo := fmt.Sprintf("%s - %s", fechaInicio.Format("02/01/2006"), fechaFin.Format("02/01/2006"))

	switch formato {
	case "pdf":
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Reporte de Consumo Eléctrico")
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 8, "Período: "+periodo)
		pdf.Ln(12)

		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(20, 20, 20)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(60, 7, "Dispositivo", "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 7, "Tipo", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 7, "Estado", "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 7, "Consumo (kWh)", "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 7, "Costo (CLP)", "1", 1, "C", true, 0, "")

		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(0, 0, 0)
		totalConsumo, totalCosto := 0.0, 0.0
		for _, d := range dispositivos {
			pdf.CellFormat(60, 6, d.Nombre, "1", 0, "L", false, 0, "")
			pdf.CellFormat(35, 6, d.Tipo, "1", 0, "L", false, 0, "")
			pdf.CellFormat(25, 6, d.Estado, "1", 0, "C", false, 0, "")
			if d.UltimaLectura != nil {
				totalConsumo += d.UltimaLectura.Energy
				totalCosto += d.UltimaLectura.Cost
				pdf.CellFormat(35, 6, fmt.Sprintf("%.3f", d.UltimaLectura.Energy), "1", 0, "R", false, 0, "")
				pdf.CellFormat(35, 6, fmt.Sprintf("$%.0f", d.UltimaLectura.Cost), "1", 1, "R", false, 0, "")
			} else {
				pdf.CellFormat(35, 6, "-", "1", 0, "C", false, 0, "")
				pdf.CellFormat(35, 6, "-", "1", 1, "C", false, 0, "")
			}
		}
		pdf.Ln(4)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 7, fmt.Sprintf("Total: %.3f kWh  |  $%.0f CLP", totalConsumo, totalCosto))

		buf := new(bytes.Buffer)
		if err := pdf.Output(buf); err != nil {
			return nil, "", "", err
		}
		return buf.Bytes(), fmt.Sprintf("consumo_%s.pdf", time.Now().Format("20060102_150405")), "application/pdf", nil

	default:
		f := excelize.NewFile()
		defer f.Close()
		sheet := "Consumo"
		idx, _ := f.NewSheet(sheet)
		f.SetActiveSheet(idx)

		headers := []string{"Dispositivo", "Tipo", "Estado", "Consumo (kWh)", "Costo (CLP)", "Período"}
		for i, h := range headers {
			f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), h)
		}
		for i, d := range dispositivos {
			r := i + 2
			f.SetCellValue(sheet, fmt.Sprintf("A%d", r), d.Nombre)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", r), d.Tipo)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", r), d.Estado)
			if d.UltimaLectura != nil {
				f.SetCellValue(sheet, fmt.Sprintf("D%d", r), d.UltimaLectura.Energy)
				f.SetCellValue(sheet, fmt.Sprintf("E%d", r), d.UltimaLectura.Cost)
			}
			f.SetCellValue(sheet, fmt.Sprintf("F%d", r), periodo)
		}
		buf := new(bytes.Buffer)
		if err := f.Write(buf); err != nil {
			return nil, "", "", err
		}
		return buf.Bytes(), fmt.Sprintf("consumo_%s.xlsx", time.Now().Format("20060102_150405")),
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	}
}
