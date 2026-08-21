package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"fintech-labs/internal/services"
	"fintech-labs/internal/utils"
)

func AdminLoanEligibilityPolicyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	savingsMultiple, err := strconv.ParseFloat(r.FormValue("savings_multiple"), 64)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid+savings+multiple", http.StatusSeeOther)
		return
	}
	shareMultiple, err := strconv.ParseFloat(r.FormValue("share_multiple"), 64)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid+share+capital+multiple", http.StatusSeeOther)
		return
	}
	minimumShares, err := strconv.ParseInt(r.FormValue("minimum_share_capital"), 10, 64)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid+minimum+share+capital", http.StatusSeeOther)
		return
	}
	err = services.SetLoanEligibilityPolicy(utils.GetSessionUser(w, r), savingsMultiple, shareMultiple, minimumShares)
	if err != nil {
		http.Redirect(w, r, "/admin?error="+strings.ReplaceAll(err.Error(), " ", "+"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin?success=Loan+eligibility+policy+updated", http.StatusSeeOther)
}
