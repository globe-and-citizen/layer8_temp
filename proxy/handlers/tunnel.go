package handlers

import (
	"fmt"
	"io"
	"net/http"
)

// Tunnel forwards the request to the service provider's backend
func Tunnel(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n\n*************")
	fmt.Println(r.Method) // >GET >POST
	fmt.Println(r.URL)    // http://localhost:5000/api/v1 => /api/v1
	//fmt.Println(io.ReadAll(r.Body)) // > [] <nil>
	fmt.Println(r)

	// get the service provider's backend url
	// host := r.Header.Get("X-Forwarded-Host")
	// scheme := r.Header.Get("X-Forwarded-Proto")
	// url := scheme + "://" + host + r.URL.Path

	backendURL := fmt.Sprintf("http://localhost:8000%s", r.URL)

	// create the request
	req, err := http.NewRequest(r.Method, backendURL, r.Body)
	if err != nil {
		fmt.Println("Error creating request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// add headers
	// req.Header.Add("X-Layer8-Proxy", "true")
	for k, v := range r.Header {
		req.Header[k] = v
	}

	// send the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Received response from 8000:", backendURL, " of code: ", res.StatusCode)

	// copy response back
	for k, v := range res.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(res.StatusCode)
	io.Copy(w, res.Body)
}
