package main
import (
"net/http"
"encoding/json"
"fmt"



)


var accounts = make(map[string]Account)
func main() {
	http.HandleFunc("/account",CreateAccount)
	fmt.Println("The app is live on http://8080/account")
	http.ListenAndServe(":8080",nil)
	
}

func CreateAccount(w http.ResponseWriter,r *http.Request){

	if r.Method !=http.MethodPost{
		http.Error(w,"Only POST is allowed",http.StatusMethodNotAllowed)
		return
	}

	
	var req Account 

	err:=json.NewDecoder(r.Body).Decode(&req)

	if err !=nil{
		http.Error(w, "Invalid request body",http.StatusBadRequest)
		return 
	}
	
fmt.Println("UserName: ",req.Username)
fmt.Println("Balance: ",req.Balance)


if req.Username==""{
	http.Error(w,"Username is required",http.StatusBadRequest)
	return 
}

if _,exists :=accounts[req.Username];exists{
	http.Error(w,"Account already exists",http.StatusBadRequest)
	return 
}

req.Balance=0
accounts[req.Username]=req

w.Header().Set("Content-Type","application/json")
json.NewEncoder(w).Encode(req)

}