package synchronization_server

import (
	"context"
	"fmt"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/storage"
)

func GetOrCreateClientID(ctx context.Context, projectName string, synchronizationClient *SynchronizationClient, stagesStorage storage.StagesStorage) (string, error) {
	clientIDRecords, err := stagesStorage.GetClientIDRecords(ctx, projectName)
	if err != nil {
		return "", err
	}

	if len(clientIDRecords) > 0 {
		res := selectOldestClientIDRecord(clientIDRecords)
		logboek.Context(ctx).Debug().LogF("GetOrCreateClientID %s selected clientID: %s\n", projectName, res.String())
		return res.ClientID, nil
	}

	newClientID, err := synchronizationClient.NewClientID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	timestampMillisec := now.Unix()*1000 + now.UnixNano()/1000_000
	rec := &storage.ClientIDRecord{ClientID: newClientID, TimestampMillisec: timestampMillisec}

	if err := stagesStorage.PostClientIDRecord(ctx, projectName, rec); err != nil {
		return "", err
	}

	// wait between posting new id and getting current id to lower probability of collision with another process posting new client-id
	time.Sleep(2 * time.Second)

	clientIDRecords, err = stagesStorage.GetClientIDRecords(ctx, projectName)
	if err != nil {
		return "", err
	}

	if len(clientIDRecords) > 0 {
		res := selectOldestClientIDRecord(clientIDRecords)
		logboek.Context(ctx).Debug().LogF("GetOrCreateClientID %s selected clientID: %s\n", projectName, res.String())
		return res.ClientID, nil
	}

	return "", fmt.Errorf("could not find clientID in storage %s after successful creation", stagesStorage.String())
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
