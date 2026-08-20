package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"fintech-labs/internal/services"
	"fintech-labs/internal/utils"
)

func AdminDistributionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	savings, err := strconv.ParseFloat(r.FormValue("savings_interest_rate"), 64)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid+savings+interest+rate", http.StatusSeeOther)
		return
	}
	dividend, err := strconv.ParseFloat(r.FormValue("share_dividend_rate"), 64)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid+share+dividend+rate", http.StatusSeeOther)
		return
	}
	savingsWithholding, _ := strconv.ParseFloat(r.FormValue("savings_withholding_rate"), 64)
	dividendWithholding, _ := strconv.ParseFloat(r.FormValue("dividend_withholding_rate"), 64)
	autoPreview := r.FormValue("auto_preview") == "on"
	err = services.SetDistributionPolicy(utils.GetSessionUser(w, r), savings, dividend, savingsWithholding, dividendWithholding, autoPreview)
	if err != nil {
		http.Redirect(w, r, "/admin?error="+strings.ReplaceAll(err.Error(), " ", "+"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin?success=Distribution+policy+updated", http.StatusSeeOther)
}

func AdminDistributionApprovalHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	runID, _ := strconv.ParseUint(r.FormValue("run_id"), 10, 64)
	err := services.ApproveDistribution(utils.GetSessionUser(w, r), uint(runID), r.FormValue("approval_reference"))
	if err != nil {
		http.Redirect(w, r, "/admin?error="+strings.ReplaceAll(err.Error(), " ", "+"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin?success=Distribution+approved", http.StatusSeeOther)
}

func AdminDistributionPreviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	run, err := services.PreviewDistribution(utils.GetSessionUser(w, r), r.FormValue("type"), r.FormValue("period"))
	if err != nil {
		http.Redirect(w, r, "/admin?error="+strings.ReplaceAll(err.Error(), " ", "+"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin?success=Distribution+preview+created+"+run.Period, http.StatusSeeOther)
}

func AdminDistributionPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	runID, _ := strconv.ParseUint(r.FormValue("run_id"), 10, 64)
	err := services.PostDistribution(utils.GetSessionUser(w, r), uint(runID))
	if err != nil {
		http.Redirect(w, r, "/admin?error="+strings.ReplaceAll(err.Error(), " ", "+"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin?success=Distribution+posted", http.StatusSeeOther)
}
