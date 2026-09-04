package helpers

import (
	"bytes"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/YugaAdiIrawan/model/report"
)

func sourceLabel(source int, noRef string) string {
	if source == 0 {
		return "UPLOAD"
	}
	// source = 1: bedakan ESUBMISSION vs EXTERNAL dari prefix no_ref
	if strings.HasPrefix(noRef, "018-") {
		return "ESUBMISSION"
	}
	return "EXTERNAL"
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeDateStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

var autoUWHeaders = []string{
	"No", "Lead ID", "No Ref", "Partner", "Name of Insurance",
	"Date of Birth", "Gender", "MCU Package", "Tenor", "UP",
	"Tanggal Order", "Source", "Status Sent CN",
}

func GenerateAutoUWExcelHTML(data []report.AutoUWReport, reportDate time.Time) ([]byte, error) {
	var buf bytes.Buffer

	title := fmt.Sprintf("Auto UW Report %s", reportDate.Format("02 January 2006"))

	// XML namespace is required so Excel recognizes the HTML as a valid workbook
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString("\n")
	buf.WriteString(`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">`)
	buf.WriteString("\n")
	buf.WriteString(`<html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:x="urn:schemas-microsoft-com:office:excel" xmlns="http://www.w3.org/1999/xhtml">`)
	buf.WriteString("\n<head>")
	buf.WriteString(`<meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>`)
	buf.WriteString("\n<!--[if gte mso 9]><xml><x:ExcelWorkbook><x:ExcelWorksheets>")
	buf.WriteString("<x:ExcelWorksheet><x:Name>Auto UW Report</x:Name>")
	buf.WriteString("<x:WorksheetOptions><x:FreezePanes/><x:FrozenNoSplit/>")
	buf.WriteString("<x:SplitHorizontal>2</x:SplitHorizontal>")
	buf.WriteString("<x:TopRowBottomPane>2</x:TopRowBottomPane>")
	buf.WriteString("</x:WorksheetOptions></x:ExcelWorksheet>")
	buf.WriteString("</x:ExcelWorksheets></x:ExcelWorkbook></xml><![endif]-->")
	buf.WriteString(`
<style>
  body { font-family: Calibri, Arial, sans-serif; font-size: 11pt; }
  table { border-collapse: collapse; width: 100%; }
  .title-row td {
    font-size: 13pt;
    font-weight: bold;
    color: #1F4E79;
    text-align: center;
    vertical-align: middle;
    padding: 6px 4px;
  }
  .header-row th {
    background-color: #1F4E79;
    color: #FFFFFF;
    font-weight: bold;
    font-size: 11pt;
    text-align: center;
    vertical-align: middle;
    border: 1px solid #FFFFFF;
    padding: 5px 8px;
    white-space: nowrap;
  }
  .data-row td {
    vertical-align: middle;
    border: 1px solid #D9D9D9;
    padding: 4px 8px;
    font-size: 10pt;
  }
  .data-row:nth-child(even) td {
    background-color: #F2F7FC;
  }
  .col-no       { width: 40px;  text-align: center; }
  .col-leadid   { width: 80px;  text-align: center; }
  .col-noref    { width: 160px; }
  .col-partner  { width: 180px; }
  .col-name     { width: 220px; }
  .col-dob      { width: 160px; text-align: center; }
  .col-gender   { width: 80px;  text-align: center; }
  .col-mcu      { width: 130px; }
  .col-tenor    { width: 80px;  text-align: center; }
  .col-up       { width: 160px; text-align: right;  mso-number-format:'\#\,\#\#0'; }
  .col-order    { width: 170px; text-align: center; }
  .col-source   { width: 180px;  text-align: center; }
  .col-status   { width: 260px; }
</style>
</head>
<body>
`)

	colCount := len(autoUWHeaders)

	buf.WriteString("<table>")

	// Title row — merged across all columns via colspan
	buf.WriteString(`<tr class="title-row">`)
	buf.WriteString(fmt.Sprintf(`<td colspan="%d">%s</td>`, colCount, html.EscapeString(title)))
	buf.WriteString("</tr>\n")

	// Header row
	colClasses := []string{
		"col-no", "col-leadid", "col-noref", "col-partner", "col-name",
		"col-dob", "col-gender", "col-mcu", "col-tenor", "col-up",
		"col-order", "col-source", "col-status",
	}
	buf.WriteString(`<tr class="header-row">`)
	for i, h := range autoUWHeaders {
		buf.WriteString(fmt.Sprintf(`<th class="%s">%s</th>`, colClasses[i], html.EscapeString(h)))
	}
	buf.WriteString("</tr>\n")

	// Data rows
	for idx, item := range data {
		statusCN := safeString(item.StatusSentCN)
		if statusCN == "" {
			statusCN = "NULL"
		}

		buf.WriteString(`<tr class="data-row">`)
		buf.WriteString(fmt.Sprintf(`<td class="col-no">%d</td>`, idx+1))
		buf.WriteString(fmt.Sprintf(`<td class="col-leadid">%d</td>`, item.LeadID))
		buf.WriteString(fmt.Sprintf(`<td class="col-noref">%s</td>`, html.EscapeString(item.NoRef)))
		buf.WriteString(fmt.Sprintf(`<td class="col-partner">%s</td>`, html.EscapeString(item.Partner)))
		buf.WriteString(fmt.Sprintf(`<td class="col-name">%s</td>`, html.EscapeString(item.NameOfInsurance)))
		buf.WriteString(fmt.Sprintf(`<td class="col-dob">%s</td>`, html.EscapeString(safeDateStr(item.DateOfBirth))))
		buf.WriteString(fmt.Sprintf(`<td class="col-gender">%s</td>`, html.EscapeString(item.Gender)))
		buf.WriteString(fmt.Sprintf(`<td class="col-mcu">%s</td>`, html.EscapeString(item.MCUPackage)))
		buf.WriteString(fmt.Sprintf(`<td class="col-tenor" x:num>%v</td>`, item.Tenor))
		buf.WriteString(fmt.Sprintf(`<td class="col-up" x:num>%v</td>`, item.UP))
		buf.WriteString(fmt.Sprintf(`<td class="col-order">%s</td>`, html.EscapeString(item.TanggalOrder.Format("2006-01-02 15:04:05"))))
		buf.WriteString(fmt.Sprintf(`<td class="col-source">%s</td>`, html.EscapeString(sourceLabel(item.Source, item.OrderID))))
		buf.WriteString(fmt.Sprintf(`<td class="col-status">%s</td>`, html.EscapeString(statusCN)))
		buf.WriteString("</tr>\n")
	}

	buf.WriteString("</table>\n</body>\n</html>")

	return buf.Bytes(), nil
}

var monthlyReportHeaders = []string{
	"No", "Name of Insurance", "Connectivity", "Status",
	"Partner", "No Ref", "MCU Package", "Tanggal Create",
}

func GenerateMonthlyReportExcelHTML(data []report.MonthlyReport, startDate, endDate time.Time) ([]byte, error) {
	var buf bytes.Buffer

	title := fmt.Sprintf("Monthly Report %s s/d %s",
		startDate.Format("02 January 2006"),
		endDate.Format("02 January 2006"),
	)

	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString("\n")
	buf.WriteString(`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">`)
	buf.WriteString("\n")
	buf.WriteString(`<html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:x="urn:schemas-microsoft-com:office:excel" xmlns="http://www.w3.org/1999/xhtml">`)
	buf.WriteString("\n<head>")
	buf.WriteString(`<meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>`)
	buf.WriteString("\n<!--[if gte mso 9]><xml><x:ExcelWorkbook><x:ExcelWorksheets>")
	buf.WriteString("<x:ExcelWorksheet><x:Name>Monthly Report</x:Name>")
	buf.WriteString("<x:WorksheetOptions><x:FreezePanes/><x:FrozenNoSplit/>")
	buf.WriteString("<x:SplitHorizontal>2</x:SplitHorizontal>")
	buf.WriteString("<x:TopRowBottomPane>2</x:TopRowBottomPane>")
	buf.WriteString("</x:WorksheetOptions></x:ExcelWorksheet>")
	buf.WriteString("</x:ExcelWorksheets></x:ExcelWorkbook></xml><![endif]-->")
	buf.WriteString(`
<style>
  body { font-family: Calibri, Arial, sans-serif; font-size: 11pt; }
  table { border-collapse: collapse; width: 100%; }
  .title-row td {
    font-size: 13pt; font-weight: bold; color: #1F4E79;
    text-align: center; vertical-align: middle; padding: 6px 4px;
  }
  .header-row th {
    background-color: #1F4E79; color: #FFFFFF; font-weight: bold;
    font-size: 11pt; text-align: center; vertical-align: middle;
    border: 1px solid #FFFFFF; padding: 5px 8px; white-space: nowrap;
  }
  .data-row td {
    vertical-align: middle; border: 1px solid #D9D9D9;
    padding: 4px 8px; font-size: 10pt;
  }
  .data-row:nth-child(even) td { background-color: #F2F7FC; }
  .col-no        { width: 40px;  text-align: center; }
  .col-name      { width: 220px; }
  .col-connect   { width: 120px; text-align: center; }
  .col-status    { width: 110px; text-align: center; }
  .col-partner   { width: 180px; }
  .col-noref     { width: 160px; }
  .col-mcu       { width: 130px; }
  .col-tanggal   { width: 170px; text-align: center; }
</style>
</head>
<body>
`)

	colCount := len(monthlyReportHeaders)
	buf.WriteString("<table>")

	buf.WriteString(`<tr class="title-row">`)
	buf.WriteString(fmt.Sprintf(`<td colspan="%d">%s</td>`, colCount, html.EscapeString(title)))
	buf.WriteString("</tr>\n")

	colClasses := []string{
		"col-no", "col-name", "col-connect", "col-status",
		"col-partner", "col-noref", "col-mcu", "col-tanggal",
	}
	buf.WriteString(`<tr class="header-row">`)
	for i, h := range monthlyReportHeaders {
		buf.WriteString(fmt.Sprintf(`<th class="%s">%s</th>`, colClasses[i], html.EscapeString(h)))
	}
	buf.WriteString("</tr>\n")

	for idx, item := range data {
		buf.WriteString(`<tr class="data-row">`)
		buf.WriteString(fmt.Sprintf(`<td class="col-no">%d</td>`, idx+1))
		buf.WriteString(fmt.Sprintf(`<td class="col-name">%s</td>`, html.EscapeString(item.NameOfInsurance)))
		buf.WriteString(fmt.Sprintf(`<td class="col-connect">%s</td>`, html.EscapeString(item.Connectivity)))
		buf.WriteString(fmt.Sprintf(`<td class="col-status">%s</td>`, html.EscapeString(item.Status)))
		buf.WriteString(fmt.Sprintf(`<td class="col-partner">%s</td>`, html.EscapeString(item.Partner)))
		buf.WriteString(fmt.Sprintf(`<td class="col-noref">%s</td>`, html.EscapeString(item.NoRef)))
		buf.WriteString(fmt.Sprintf(`<td class="col-mcu">%s</td>`, html.EscapeString(item.MCUPackage)))
		buf.WriteString(fmt.Sprintf(`<td class="col-tanggal">%s</td>`, item.TanggalCreate.Format("2006-01-02 15:04:05")))
		buf.WriteString("</tr>\n")
	}

	buf.WriteString("</table>\n</body>\n</html>")
	return buf.Bytes(), nil
}
