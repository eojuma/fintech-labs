package handlers

import (
	"encoding/csv"
	"fintech-labs/backend/services"
	"fintech-labs/backend/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/jung-kurt/gofpdf"
)

func DownloadStatementHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	accountNumber := r.URL.Query().Get("account")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	format := r.URL.Query().Get("format") // "pdf" or "csv"

	if accountNumber == "" || fromStr == "" || toStr == "" {
		http.Redirect(w, r, "/dashboard?error=Account+and+date+range+required", http.StatusSeeOther)
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid+from+date", http.StatusSeeOther)
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid+to+date", http.StatusSeeOther)
		return
	}

	if to.Before(from) {
		http.Redirect(w, r, "/dashboard?error=End+date+must+be+after+start+date", http.StatusSeeOther)
		return
	}

	data, err := services.GenerateStatement(username, accountNumber, from, to)
	if err != nil {
		errMsg := "Failed+to+generate+statement"
		http.Redirect(w, r, "/dashboard?error="+errMsg, http.StatusSeeOther)
		return
	}

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="statement_%s_%s_%s.csv"`, accountNumber, fromStr, toStr))

		writer := csv.NewWriter(w)
		defer writer.Flush()

		writer.Write([]string{"African Vault — Account Statement"})
		writer.Write([]string{"Account Holder", data.AccountHolderName})
		writer.Write([]string{"Account Number", data.AccountNumber})
		writer.Write([]string{"Period", fmt.Sprintf("%s to %s", fromStr, toStr)})
		writer.Write([]string{"Opening Balance", fmt.Sprintf("KES %d", data.OpeningBalance)})
		writer.Write([]string{})
		writer.Write([]string{"Date", "Type", "Amount (KES)", "Balance (KES)"})

		for _, tx := range data.Transactions {
			writer.Write([]string{
				tx.CreatedAt.Format("02 Jan 2006 15:04"),
				tx.Type,
				fmt.Sprintf("%d", tx.Amount),
				fmt.Sprintf("%d", tx.Balance),
			})
		}

		writer.Write([]string{})
		writer.Write([]string{"Closing Balance", fmt.Sprintf("KES %d", data.ClosingBalance)})
		return
	}

	// Default to PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "African Vault — Account Statement")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(60, 8, "Account Holder:")
	pdf.Cell(0, 8, data.AccountHolderName)
	pdf.Ln(8)
	pdf.Cell(60, 8, "Account Number:")
	pdf.Cell(0, 8, data.AccountNumber)
	pdf.Ln(8)
	pdf.Cell(60, 8, "Period:")
	pdf.Cell(0, 8, fmt.Sprintf("%s to %s", fromStr, toStr))
	pdf.Ln(8)
	pdf.Cell(60, 8, "Opening Balance:")
	pdf.Cell(0, 8, fmt.Sprintf("KES %d", data.OpeningBalance))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(0, 74, 153)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(45, 8, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 8, "Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 8, "Amount (KES)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 8, "Balance (KES)", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)
	fill := false
	for _, tx := range data.Transactions {
		if fill {
			pdf.SetFillColor(230, 240, 255)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.CellFormat(45, 7, tx.CreatedAt.Format("02 Jan 2006 15:04"), "1", 0, "L", true, 0, "")
		pdf.CellFormat(45, 7, tx.Type, "1", 0, "L", true, 0, "")
		pdf.CellFormat(45, 7, fmt.Sprintf("%d", tx.Amount), "1", 0, "R", true, 0, "")
		pdf.CellFormat(45, 7, fmt.Sprintf("%d", tx.Balance), "1", 1, "R", true, 0, "")
		fill = !fill
	}

	pdf.Ln(6)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(60, 8, "Closing Balance:")
	pdf.Cell(0, 8, fmt.Sprintf("KES %d", data.ClosingBalance))

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="statement_%s_%s_%s.pdf"`, accountNumber, fromStr, toStr))
	pdf.Output(w)
}