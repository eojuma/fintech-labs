package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"fintech-labs/internal/services"
	"fintech-labs/internal/utils"
)

func AdminShareContributionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	adminUsername := utils.GetSessionUser(w, r)
	amount, err := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid+share+contribution+amount", http.StatusSeeOther)
		return
	}
	contribution, err := services.RecordShareContribution(adminUsername, r.FormValue("username"), amount, r.FormValue("note"))
	if err != nil {
		http.Redirect(w, r, "/admin?error="+strings.ReplaceAll(err.Error(), " ", "+"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin?success=Share+contribution+recorded+with+reference+"+contribution.ReferenceNumber, http.StatusSeeOther)
}

func AdminShareRedemptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	amount, err := strconv.ParseInt(r.FormValue("amount"), 10, 64)
	if err == nil {
		_, err = services.RedeemShareCapital(utils.GetSessionUser(w, r), r.FormValue("username"), amount, r.FormValue("reason"))
	}
	if err != nil {
		http.Redirect(w, r, "/admin?error="+strings.ReplaceAll(err.Error(), " ", "+"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin?success=Share+capital+redeemed", http.StatusSeeOther)
}
