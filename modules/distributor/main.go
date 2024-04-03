package distributor

import (
	"RediDB/modules/config"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

type distributor struct {
	sync.RWMutex
	Data map[string]distributorData
}

type distributorData struct {
	ID          string
	DestroyTime int64

	Documents []map[string]interface{}
}

var Distributor = &distributor{Data: make(map[string]distributorData)}
var _config = config.Get().Distribute

var (
	DestroyTimeSetter int64 = 10
	GiveDataFromCall        = _config.GiveMax

	errNotFound = errors.New("no distribute found")
)

func init() {
	go func() {
		ticker := time.NewTicker(time.Duration(DestroyTimeSetter+1) * time.Second)

		for range ticker.C {
			Distributor.Lock()
			for _, distrubute := range Distributor.Data {
				if len(distrubute.Documents) == 0 || distrubute.DestroyTime < getTimeMili() {
					deleteDistribute(distrubute.ID, false)
					ticker.Stop()
				}
			}
			Distributor.Unlock()
		}
	}()
}

func getDistribute(ID string) (distributorData, error) {
	Distributor.Lock()

	data := Distributor.Data[ID]
	Distributor.Unlock()

	if data.DestroyTime == 0 {
		return data, errNotFound
	}

	return data, nil
}

func overwriteDistribute(ID string, data distributorData) {
	Distributor.Lock()

	Distributor.Data[ID] = data

	Distributor.Unlock()
}

func deleteDistribute(ID string, lock bool) {
	if lock {
		Distributor.Lock()
		defer Distributor.Unlock()
	}

	delete(Distributor.Data, ID)
}

func getTimeMili() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func generateID(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
