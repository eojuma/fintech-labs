package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"fintech-labs/internal/services"
	"fintech-labs/internal/utils"
)

func ApplyForLoanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	principal, err := strconv.ParseInt(r.FormValue("principal"), 10, 64)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid+loan+principal", http.StatusSeeOther)
		return
	}
	term, err := strconv.Atoi(r.FormValue("term_months"))
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid+loan+term", http.StatusSeeOther)
		return
	}
	loan, err := services.ApplyForLoan(utils.GetSessionUser(w, r), r.FormValue("purpose"), principal, term)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error="+strings.ReplaceAll(err.Error(), " ", "+"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/dashboard?success=Loan+application+submitted+with+reference+"+loan.ReferenceNumber, http.StatusSeeOther)
}

func AdminLoanDecisionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	loanID, _ := strconv.ParseUint(r.FormValue("loan_id"), 10, 64)
	principal, _ := strconv.ParseInt(r.FormValue("approved_principal"), 10, 64)
	total, _ := strconv.ParseInt(r.FormValue("total_repayable"), 10, 64)
	err := services.DecideLoan(utils.GetSessionUser(w, r), uint(loanID), r.FormValue("action"), principal, total, r.FormValue("note"))
	redirectLoanResult(w, r, err, "Loan decision recorded")
}

func AdminLoanDisbursementHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	loanID, _ := strconv.ParseUint(r.FormValue("loan_id"), 10, 64)
	err := services.DisburseLoan(utils.GetSessionUser(w, r), uint(loanID))
	redirectLoanResult(w, r, err, "Loan disbursed")
}

func AdminLoanRepaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	loanID, _ := strconv.ParseUint(r.FormValue("loan_id"), 10, 64)
	amount, err := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	if err == nil {
		err = services.RecordLoanRepayment(utils.GetSessionUser(w, r), uint(loanID), amount)
	}
	redirectLoanResult(w, r, err, "Loan repayment recorded")
}

func redirectLoanResult(w http.ResponseWriter, r *http.Request, err error, success string) {
	if err != nil {
		http.Redirect(w, r, "/admin?error="+strings.ReplaceAll(err.Error(), " ", "+"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin?success="+strings.ReplaceAll(success, " ", "+"), http.StatusSeeOther)
}
