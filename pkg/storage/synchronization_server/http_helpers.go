package synchronization_server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func PerformPost(client *http.Client, url string, request, response interface{}) error {
	reqBodyData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("unable to marshal request data: %s", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyData))
	if err != nil {
		return fmt.Errorf("unable to create POST request for %q: %s", url, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error requesting url %q: %s", url, err)
	}

	defer resp.Body.Close()
	respBodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response of %q request: %s", url, err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("got bad response %s by url %q request:\n%s", resp.Status, url, string(respBodyData))
	}

	if err := json.Unmarshal(respBodyData, response); err != nil {
		return fmt.Errorf("unable to unmarshal json body by url %q request: %s", url, err)
	}

	return nil
}

func HandleRequest(w http.ResponseWriter, r *http.Request, request, response interface{}, actionFunc func()) {
	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, fmt.Sprintf("unable to unmarshal request json: %s", err), http.StatusBadRequest)
			return
		}
	}

	actionFunc()

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
