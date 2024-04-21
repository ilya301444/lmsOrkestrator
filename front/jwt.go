package front

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (d *Data) registrUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("reg")
	login := r.FormValue("login")
	password := r.FormValue("password")
	fmt.Println(login, password)
	if login == "" || password == "" {
		http.Error(w, "error in you reaquest!", http.StatusBadRequest)
		return
	}
	if _, ok := d.User[login]; ok {
		http.Error(w, "error you already autorizated", http.StatusUnauthorized)
		return
	}
	d.User[login] = User{Login: login, Pass: password}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Youre autorized in system!!"))
}

func (d *Data) login(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	password := r.FormValue("password")
	if login == "" || password == "" {
		http.Error(w, "error in you reaquest!", http.StatusBadRequest)
		return
	}
	us, ok := d.User[login]
	if !ok {
		http.Error(w, "error you unoutorized", http.StatusUnauthorized)
		return
	}
	if us.Pass != password {
		http.Error(w, "error bad password", http.StatusUnauthorized)
		return
	}
	token := newJwtTocken(login)
	http.SetCookie(w, &http.Cookie{
		Name:     "jwtToken",
		Value:    token,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Write([]byte("cookie set!"))
}

func (d *Data) pageRegistratration(w http.ResponseWriter, r *http.Request) {
	err := data.cashe["registr"].Execute(w, nil)
	PrintEr(err)
}

func (d *Data) pageLogin(w http.ResponseWriter, r *http.Request) {
	err := data.cashe["login"].Execute(w, nil)
	PrintEr(err)
}

const secret = "secret"

func newJwtTocken(name string) string {
	now := time.Now()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": name,
		"nbf":  now.Unix(),
		"exp":  now.Add(2400 * time.Hour).Unix(),
		"iat":  now.Unix(),
	})

	tStr, err := t.SignedString([]byte(secret))
	if err != nil {
		return ""
	}
	return tStr
}

func checkValidJwt(jwtStr string) bool {
	tockenfromStr, err := jwt.Parse(jwtStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return false, fmt.Errorf("unexpected error!! %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		Log(err)
		return false
	}
	if !tockenfromStr.Valid {
		fmt.Println("tocken invalid 1")
		return false
	}
	return true
}
