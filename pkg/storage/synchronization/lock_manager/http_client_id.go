package lock_manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/synchronization/server"
)

func GetHttpClientID(ctx context.Context, projectName, serverAddress string, stagesStorage storage.StagesStorage) (string, error) {
	// Try to get clientID from storage.
	{
		clientID, err := getClientIDFromStorage(ctx, projectName, stagesStorage)
		if err != nil {
			return "", err
		}

		if clientID != "" {
			logboek.Context(ctx).Debug().LogF("getClientIDFromStorage %s selected clientID: %s\n", projectName, clientID)
			return clientID, nil
		}
	}

	// Create new clientID and post it to storage.
	{
		clientID, err := newClientID(serverAddress, &http.Client{})
		if err != nil {
			return "", err
		}

		now := time.Now()
		timestampMillisec := now.Unix()*1000 + now.UnixNano()/1000_000
		rec := &storage.ClientIDRecord{ClientID: clientID, TimestampMillisec: timestampMillisec}

		if err := stagesStorage.PostClientIDRecord(ctx, projectName, rec); err != nil {
			return "", err
		}
	}

	// Retry getting clientID to account for the possibility that another process created it before.
	{
		time.Sleep(2 * time.Second)

		clientID, err := getClientIDFromStorage(ctx, projectName, stagesStorage)
		if err != nil {
			return "", err
		}

		if clientID != "" {
			logboek.Context(ctx).Debug().LogF("getClientIDFromStorage %s selected clientID: %s\n", projectName, clientID)
			return clientID, nil
		}
	}

	return "", fmt.Errorf("could not find clientID in storage %s after successful creation", stagesStorage.String())
}

func getClientIDFromStorage(ctx context.Context, projectName string, stagesStorage storage.StagesStorage) (string, error) {
	clientIDRecords, err := stagesStorage.GetClientIDRecords(ctx, projectName)
	if err != nil {
		return "", err
	}

	if len(clientIDRecords) > 0 {
		res := selectOldestClientIDRecord(clientIDRecords)
		logboek.Context(ctx).Debug().LogF("GetOrCreateHttpClientID %s selected clientID: %s\n", projectName, res.String())
		return res.ClientID, nil
	}

	return "", nil
}

func selectOldestClientIDRecord(records []*storage.ClientIDRecord) *storage.ClientIDRecord {
	var foundRec *storage.ClientIDRecord
	for _, rec := range records {
		if foundRec == nil || rec.TimestampMillisec < foundRec.TimestampMillisec {
			foundRec = rec
		}
	}
	return foundRec
}

func newClientID(url string, httpClient *http.Client) (string, error) {
	request := server.NewClientIDRequest{}
	response := server.NewClientIDResponse{}
	if err := performPost(httpClient, fmt.Sprintf("%s/%s", url, "new-client-id"), request, &response); err != nil {
		return "", err
	}
	return response.ClientID, response.Err.Error
}

func performPost(client *http.Client, url string, request, response interface{}) error {
	reqBodyData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("unable to marshal request data: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyData))
	if err != nil {
		return fmt.Errorf("unable to create POST request for %q: %w", url, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error requesting url %q: %w", url, err)
	}

	defer resp.Body.Close()
	respBodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response of %q request: %w", url, err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("got bad response %s by url %q request:\n%s", resp.Status, url, string(respBodyData))
	}

	if err := json.Unmarshal(respBodyData, response); err != nil {
		return fmt.Errorf("unable to unmarshal json body by url %q request: %w", url, err)
	}

	return nil
}
