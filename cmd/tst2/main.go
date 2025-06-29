package main

import (
	"encoding/json"
	"log"
	"net/url"
)

func main() {
	values, err := url.ParseQuery("pile_field_id=1&project_id=1&recorded_by=%D0%92%D0%B0%D1%81%D1%8F&status=30")
	//values, err := url.Parse("http://localhost:8080/getpiles?pile_field_id=1&project_id=1&recorded_by=%D0%92%D0%B0%D1%81%D1%8F&status=30")
	if err != nil {
		log.Fatal("parse:", err)
	}
	println("values:", values.Get("recorded_by"))
	jsonData, err := json.Marshal(values)
	if err != nil {
		log.Fatal("marshal:", err)
	}
	println("result:", jsonData)
}
