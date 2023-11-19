package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"globe-and-citizen/layer8/utils"
	"io"
	"net/http"
	"os"
)

// Tunnel forwards the request to the service provider's backend
func InitTunnel(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n\n*************")
	fmt.Println(r.Method) // > GET  | > POST
	fmt.Println(r.URL)    // (http://localhost:5000/api/v1 ) > /api/v1

	mpJWT, err := utils.GenerateStandardToken(os.Getenv("MP_123_SECRET_KEY"))
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	backendURL := fmt.Sprintf("http://localhost:8000%s", r.URL)

	// create the request
	req, err := http.NewRequest(r.Method, backendURL, r.Body)
	if err != nil {
		fmt.Println("Error creating request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// add headers
	for k, v := range r.Header {
		req.Header[k] = v
	}

	req.Header["mp_JWT"] = []string{mpJWT}

	// send the request
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("Error sending request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Make a buffer to hold response body
	var resBodyTemp bytes.Buffer

	// Copy the response body to buffer
	resBodyTemp.ReadFrom(res.Body)

	// Convert resBodyTemp to []byte

	resBodyTempBytes := []byte(resBodyTemp.String())

	fmt.Println("resBodyTempBytes: ", string(resBodyTempBytes))

	// Make a copy of the response body to send back to client
	res.Body = io.NopCloser(bytes.NewBuffer(resBodyTemp.Bytes()))

	fmt.Println("\nReceived response from 8000:", backendURL, " of code: ", res.StatusCode)

	// copy response back
	// for k, v := range res.Header {
	// 	w.Header()[k] = v
	// 	fmt.Println("header pairs from SP: ", k, v)
	// }

	upJWT, err := utils.GenerateStandardToken(os.Getenv("UP_999_SECRET_KEY"))
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header()["up_JWT"] = []string{upJWT}
	w.Header().Set("up_JWT", upJWT)
	// fmt.Println("w.Headers (Going back to client): ", w.Header())
	// Headers not being sent back to client for some reason...

	server_pubKeyECDH, err := utils.B64ToJWK(string(resBodyTempBytes))
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	fmt.Println("server_pubKeyECDH: ", server_pubKeyECDH)

	// // Make a json response of server_pubKeyECDH and up_JWT and send it back to client
	data := map[string]interface{}{
		"server_pubKeyECDH": server_pubKeyECDH,
		"up_JWT":            upJWT,
	}

	fmt.Println("data (Going back to client): ", data)

	datatoSend, err := json.Marshal(&data)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataToSendB64 := base64.URLEncoding.EncodeToString(datatoSend)

	// io.Copy(w, bytes.NewBufferString(dataToSendB64))

	// io.Copy(w, strings.NewReader(dataToSendB64))

	// io.Copy(w, dataIoReader)

	// io.CopyBuffer(w, bytes.NewBufferString(dataToSendB64), []byte(dataToSendB64))

	w.Write([]byte(dataToSendB64))

}

func Tunnel(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n\n*************")
	fmt.Println(r.Method) // > GET  | > POST
	fmt.Println(r.URL)    // (http://localhost:5000/api/v1 ) > /api/v1

	backendURL := fmt.Sprintf("http://localhost:8000%s", r.URL)

	// create the request
	req, err := http.NewRequest(r.Method, backendURL, r.Body)
	if err != nil {
		fmt.Println("Error creating request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// add headers
	for k, v := range r.Header {
		req.Header[k] = v
		fmt.Println("header pairs from client (Interceptor): ", k, v)
	}

	// Get up_JWT from request header and verify it
	upJWT := r.Header.Get("up_JWT")

	fmt.Println("up_JWT Check 1: ", upJWT)

	// _, err = utils.VerifyStandardToken(upJWT, os.Getenv("UP_999_SECRET_KEY"))
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	http.Error(w, err.Error(), http.StatusUnauthorized)
	// 	return
	// }

	// send the request
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("Error sending request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("\nReceived response from 8000:", backendURL, " of code: ", res.StatusCode)

	// Get mp_JWT from response header and verify it
	// mpJWT := res.Header.Get("mp_JWT")

	// _, err = utils.VerifyStandardToken(mpJWT, os.Getenv("MP_123_SECRET_KEY"))
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	http.Error(w, err.Error(), http.StatusUnauthorized)
	// 	return
	// }

	// copy response back
	for k, v := range res.Header {
		w.Header()[k] = v
		//fmt.Println("header pairs from SP: ", k, v)
	}

	w.Header()["setme"] = []string{"string"}
	w.Header().Add("ME TOO?", "DO IT!")
	fmt.Println("w.Headers: ", w.Header())
	//w.WriteHeader(res.StatusCode)
	io.Copy(w, res.Body)

	fmt.Println("w.Headers 2: ", w.Header())
}
