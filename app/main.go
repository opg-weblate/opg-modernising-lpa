package main

import (
	"fmt"
	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/fake"
	"html/template"
	"log"
	"net/http"
)

func Hello() string {
	return "Hello, world!"
}

type PageData struct {
	WebDir      string
	Prefix      string
	PrefixAsset string
}

func home(w http.ResponseWriter, r *http.Request) {
	webDir := env.Get("WEB_DIR", "web")
	prefix := env.Get("PREFIX", "")

	data := PageData{
		WebDir: webDir,
		Prefix: prefix,
	}

	files := []string{
		fmt.Sprintf("%s/template/home.gohtml", webDir),
		fmt.Sprintf("%s/template/layout/base.gohtml", webDir),
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		log.Fatal(err)
	}

	err = ts.ExecuteTemplate(w, "base", data)

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	fmt.Println(fake.GoodBye())

	mux := http.NewServeMux()
	mux.HandleFunc("/home", home)

	err := http.ListenAndServe(":5000", mux)

	if err != nil {
		log.Fatal(err)
	}
}
