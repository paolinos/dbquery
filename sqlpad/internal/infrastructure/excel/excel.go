package excel

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"

	"github.com/dbquery/dbquery/internal/core/models"
)

// Parser handles Excel file parsing.
type Parser struct{}

// NewParser creates a new Excel parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile reads an Excel file and returns all sheets with their data.
func (p *Parser) ParseFile(filePath string) ([]models.ExcelSheet, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file %s: %w", filePath, err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	var result []models.ExcelSheet
	for _, sheetName := range sheets {
		sheet, err := p.parseSheet(f, sheetName)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sheet %s: %w", sheetName, err)
		}
		if sheet != nil {
			result = append(result, *sheet)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no data found in any sheet")
	}

	return result, nil
}

// ParseReader reads an Excel file from a byte slice (useful for uploaded files).
func (p *Parser) ParseReader(data []byte, fileName string) ([]models.ExcelSheet, error) {
	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		// Try opening from bytes directly
		f, err = excelize.OpenReader(strings.NewReader(string(data)))
		if err != nil {
			return nil, fmt.Errorf("failed to open Excel file %s: %w", fileName, err)
		}
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	var result []models.ExcelSheet
	for _, sheetName := range sheets {
		sheet, err := p.parseSheet(f, sheetName)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sheet %s: %w", sheetName, err)
		}
		if sheet != nil {
			result = append(result, *sheet)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no data found in any sheet")
	}

	return result, nil
}

func (p *Parser) parseSheet(f *excelize.File, sheetName string) (*models.ExcelSheet, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet %s: %w", sheetName, err)
	}

	if len(rows) < 2 {
		// Need at least a header row and one data row
		return nil, nil
	}

	// First row is headers
	headers := make([]string, len(rows[0]))
	for i, h := range rows[0] {
		headers[i] = strings.TrimSpace(h)
		if headers[i] == "" {
			headers[i] = fmt.Sprintf("column_%d", i+1)
		}
	}

	// Data rows
	var dataRows [][]string
	for _, row := range rows[1:] {
		// Skip completely empty rows
		allEmpty := true
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue
		}

		dataRow := make([]string, len(headers))
		for i, cell := range row {
			if i < len(headers) {
				dataRow[i] = cell
			}
		}
		// Fill remaining columns with empty string
		for i := len(row); i < len(headers); i++ {
			dataRow[i] = ""
		}
		dataRows = append(dataRows, dataRow)
	}

	if len(dataRows) == 0 {
		return nil, nil
	}

	return &models.ExcelSheet{
		Name:    sheetName,
		Headers: headers,
		Rows:    dataRows,
	}, nil
}

// IsExcelFile checks if the given filename has an Excel extension.
func IsExcelFile(fileName string) bool {
	lower := strings.ToLower(fileName)
	return strings.HasSuffix(lower, ".xlsx") ||
		strings.HasSuffix(lower, ".xlsm") ||
		strings.HasSuffix(lower, ".xltx") ||
		strings.HasSuffix(lower, ".xltm")
}

// NormalizeSheetName converts a sheet name to a valid SQL table name.
func NormalizeSheetName(name string) string {
	name = strings.TrimSpace(name)
	var result strings.Builder
	for i, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '.' || r == '/' || r == '\\' || r == '(' || r == ')' {
			if i > 0 && result.String()[result.Len()-1] != '_' {
				result.WriteRune('_')
			}
		} else if r == '_' {
			result.WriteRune('_')
		}
	}
	name = strings.TrimLeft(name, "_")
	name = strings.TrimRight(name, "_")
	if name == "" {
		name = "sheet"
	}
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "s_" + name
	}
	return strings.ToLower(name)
}
