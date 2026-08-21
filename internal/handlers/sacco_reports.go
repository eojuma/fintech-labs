package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"fintech-labs/internal/db"
	"fintech-labs/internal/models"
)

func SaccoReportHandler(w http.ResponseWriter, r *http.Request) {
	reportType := r.URL.Query().Get("type")
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-report.csv", reportType))
	writer := csv.NewWriter(w)
	defer writer.Flush()
	switch reportType {
	case "loans":
		writer.Write([]string{"Reference", "Member", "Status", "Requested", "Approved", "Outstanding Principal", "Outstanding Interest", "Term Months", "Disbursed At"})
		var loans []models.Loan
		db.DB.Preload("User").Order("created_at asc").Find(&loans)
		for _, loan := range loans {
			disbursed := ""
			if loan.DisbursedAt != nil {
				disbursed = loan.DisbursedAt.Format("2006-01-02")
			}
			writer.Write([]string{loan.ReferenceNumber, loan.User.Username, loan.Status, strconv.FormatInt(loan.RequestedPrincipal, 10), strconv.FormatInt(loan.ApprovedPrincipal, 10), strconv.FormatInt(loan.OutstandingPrincipal, 10), strconv.FormatInt(loan.OutstandingInterest, 10), strconv.Itoa(loan.TermMonths), disbursed})
		}
	case "shares":
		writer.Write([]string{"Member", "Balance", "Active", "Updated At"})
		var shares []models.ShareCapital
		db.DB.Preload("User").Order("user_id asc").Find(&shares)
		for _, share := range shares {
			writer.Write([]string{share.User.Username, strconv.FormatInt(share.Balance, 10), strconv.FormatBool(share.Active), share.UpdatedAt.Format("2006-01-02 15:04")})
		}
	case "distributions":
		writer.Write([]string{"Type", "Period", "Member", "Basis", "Gross", "Withholding", "Net", "Status", "Approval Reference"})
		var allocations []models.DistributionAllocation
		db.DB.Preload("User").Order("distribution_run_id asc").Find(&allocations)
		for _, allocation := range allocations {
			var run models.DistributionRun
			db.DB.First(&run, allocation.DistributionRunID)
			writer.Write([]string{run.Type, run.Period, allocation.User.Username, strconv.FormatInt(allocation.BasisAmount, 10), strconv.FormatInt(allocation.GrossAmount, 10), strconv.FormatInt(allocation.WithholdingAmount, 10), strconv.FormatInt(allocation.Amount, 10), allocation.Status, run.ApprovalReference})
		}
	default:
		http.Error(w, "unknown report type", http.StatusBadRequest)
	}
}
