package distributor

func Set(data []map[string]interface{}) string {
	Distributor.Lock()

	id := generateID(15)
	distributorData := distributorData{
		ID:          id,
		DestroyTime: getTimeMili() + DestroyTimeSetter,
		Documents:   data,
	}

	Distributor.Data[id] = distributorData
	Distributor.Unlock()

	return id
}

func GetData(ID string) ([]map[string]interface{}, int, error) {
	data, err := getDistribute(ID)
	if err != nil {
		return nil, 0, err
	}

	if len(data.Documents) == 0 {
		return nil, 0, errNotFound
	}

	documents := []map[string]interface{}{}
	data.DestroyTime = getTimeMili() + DestroyTimeSetter

	giveDataFromCall := GiveDataFromCall
	if giveDataFromCall > len(data.Documents) {
		giveDataFromCall = len(data.Documents)
	}

	for i := 0; i < giveDataFromCall; i++ {
		documents = append(documents, data.Documents[i])
	}

	data.Documents = data.Documents[giveDataFromCall:]

	overwriteDistribute(ID, data)
	return documents, len(data.Documents), nil
}
