package jsondecodebody

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang/gddo/httputil/header"
)

type MalformedRequest struct {
	Status int
	Msg    string
}

func (mr *MalformedRequest) Error() string {
	return mr.Msg
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	fmt.Println("1")
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			Msg := "Content-Type header is not application/json"
			return &MalformedRequest{Status: http.StatusUnsupportedMediaType, Msg: Msg}
		}
	}
	fmt.Println("2")

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	fmt.Println("3")
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	fmt.Println("4")

	err := dec.Decode(&dst)
	fmt.Println("5")
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		fmt.Println("6")
		switch {
		case errors.As(err, &syntaxError):
			Msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: Msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			Msg := "Request body contains badly-formed JSON"
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: Msg}

		case errors.As(err, &unmarshalTypeError):
			Msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: Msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			Msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: Msg}

		case errors.Is(err, io.EOF):
			Msg := "Request body must not be empty"
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: Msg}

		case err.Error() == "http: request body too large":
			Msg := "Request body must not be larger than 1MB"
			return &MalformedRequest{Status: http.StatusRequestEntityTooLarge, Msg: Msg}

		default:
			return err
		}
	}

	fmt.Println("7")
	err = dec.Decode(&struct{}{})
	fmt.Println("8")
	if err != io.EOF {
		Msg := "Request body must only contain a single JSON object"
		return &MalformedRequest{Status: http.StatusBadRequest, Msg: Msg}
	}
	return nil
}
