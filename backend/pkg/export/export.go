package export

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/xuri/excelize/v2"
)

type Table struct {
	SheetName string
	Headers   []string
	Rows      [][]string
}

func ToCSV(table Table) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(table.Headers); err != nil {
		return nil, err
	}
	for _, row := range table.Rows {
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func ToExcel(table Table) ([]byte, error) {
	f := excelize.NewFile()
	sheet := table.SheetName
	if sheet == "" {
		sheet = "Report"
	}
	defaultSheet := f.GetSheetName(0)
	if err := f.SetSheetName(defaultSheet, sheet); err != nil {
		return nil, err
	}
	for i, h := range table.Headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return nil, err
		}
	}
	for r, row := range table.Rows {
		for c, val := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			if err := f.SetCellValue(sheet, cell, val); err != nil {
				return nil, err
			}
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Filename(base, format string) string {
	switch format {
	case "xlsx", "excel":
		return fmt.Sprintf("%s.xlsx", base)
	default:
		return fmt.Sprintf("%s.csv", base)
	}
}

func ContentType(format string) string {
	switch format {
	case "xlsx", "excel":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "text/csv"
	}
}
