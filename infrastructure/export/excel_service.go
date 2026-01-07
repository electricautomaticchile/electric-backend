package export

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

type ExcelService struct{}

func NewExcelService() *ExcelService {
	return &ExcelService{}
}

func (s *ExcelService) ExportarClientes(data []map[string]interface{}) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Clientes"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	headers := []string{"Nombre", "Correo", "Teléfono", "Dirección", "Ciudad", "Tipo", "Estado", "Fecha Registro"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#F0F0F0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), style)

	for i, row := range data {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), row["nombre"])
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), row["correo"])
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), row["telefono"])
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), row["direccion"])
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), row["ciudad"])
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), row["tipoCliente"])
		
		estado := "Inactivo"
		if activo, ok := row["activo"].(bool); ok && activo {
			estado = "Activo"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), estado)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), row["fechaRegistro"])
	}

	for i := 0; i < len(headers); i++ {
		col := fmt.Sprintf("%c", 'A'+i)
		f.SetColWidth(sheetName, col, col, 20)
	}

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (s *ExcelService) ExportarDispositivos(data []map[string]interface{}) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Dispositivos"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	headers := []string{"ID Dispositivo", "Tipo", "Modelo", "Cliente", "Estado", "Consumo Actual", "Fecha Instalación"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#F0F0F0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), style)

	for i, row := range data {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), row["idDispositivo"])
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), row["tipo"])
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), row["modelo"])
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), row["cliente"])
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), row["estado"])
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), row["consumoActual"])
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), row["fechaInstalacion"])
	}

	for i := 0; i < len(headers); i++ {
		col := fmt.Sprintf("%c", 'A'+i)
		f.SetColWidth(sheetName, col, col, 20)
	}

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (s *ExcelService) ExportarAlertas(data []map[string]interface{}) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Alertas"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	headers := []string{"Tipo", "Mensaje", "Dispositivo", "Estado", "Fecha Creación", "Fecha Resolución"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#F0F0F0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), style)

	for i, row := range data {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), row["tipo"])
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), row["mensaje"])
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), row["dispositivo"])
		
		estado := "Pendiente"
		if resuelta, ok := row["resuelta"].(bool); ok && resuelta {
			estado = "Resuelta"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), estado)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), row["fechaCreacion"])
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), row["fechaResolucion"])
	}

	for i := 0; i < len(headers); i++ {
		col := fmt.Sprintf("%c", 'A'+i)
		f.SetColWidth(sheetName, col, col, 20)
	}

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (s *ExcelService) ExportarBoletas(data []map[string]interface{}) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Boletas"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	headers := []string{"Número", "Cliente", "Período", "Consumo (kWh)", "Monto", "Estado", "Fecha Emisión", "Fecha Vencimiento"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#F0F0F0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), style)

	for i, row := range data {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), row["numero"])
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), row["cliente"])
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), row["periodo"])
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), row["consumo"])
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), row["monto"])
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), row["estado"])
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), row["fechaEmision"])
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), row["fechaVencimiento"])
	}

	for i := 0; i < len(headers); i++ {
		col := fmt.Sprintf("%c", 'A'+i)
		f.SetColWidth(sheetName, col, col, 18)
	}

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
