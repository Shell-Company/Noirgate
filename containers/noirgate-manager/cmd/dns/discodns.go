package dns

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	config "noirgate/config"
)

func AddNoirgateRecord(NoirgateID string, IPAddress string) {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DisableKeepAlives = false
	httpClient := http.Client{Transport: t}
	defer httpClient.CloseIdleConnections()
	recordTypes := []string{"AAAA", "A"}
	sandboxDomain := strings.Split(config.SandboxDomain, ".")
	for _, recordType := range recordTypes {
		httpRequestURL := fmt.Sprintf("http://%s:2379/v2/keys/%s/%s/%s/%s/.%v", config.NoirgateETCDHost, sandboxDomain[1], sandboxDomain[0], config.SandboxSubDomain, NoirgateID, recordType)
		params := url.Values{}
		params.Add("value", IPAddress)
		body := strings.NewReader(params.Encode())

		req, err := http.NewRequest("PUT", httpRequestURL, body)
		if err != nil {
			// handle err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		httpResponse, err := httpClient.Do(req)
		responseBytes, _ := ioutil.ReadAll(httpResponse.Body)
		log.Println("Adding DNS Records", string(responseBytes))
	}

}
func DeleteNoirgateRecord(NoirgateID string) {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DisableKeepAlives = false
	httpClient := http.Client{Transport: t}
	defer httpClient.CloseIdleConnections()

	recordTypes := []string{"AAAA", "A"}
	sandboxDomain := strings.Split(config.SandboxDomain, ".")
	for _, recordType := range recordTypes {
		httpRequestURL := fmt.Sprintf("http://%s:2379/v2/keys/%s/%s/%s/%s/.%v", config.NoirgateETCDHost, sandboxDomain[1], sandboxDomain[0], config.SandboxSubDomain, NoirgateID, recordType)
		req, err := http.NewRequest("DELETE", httpRequestURL, nil)
		if err != nil {
			// handle err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		httpResponse, err := httpClient.Do(req)
		responseBytes, _ := ioutil.ReadAll(httpResponse.Body)
		log.Println("Removing DNS Records", string(responseBytes))

	}

}
