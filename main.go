package modbus_configurator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type TaskResponse struct {
	ID string `json:"id"`
	// Другие поля в соответствии с структурой ответа
}

func main() {
	// ModbusRequest := ModbusRequest{
	// 	RequestType: "Task",
	// 	}
	// Структура передается в другом месте, пример логики общения с api influxdb
	}
	AddInfluxTaskForDevice(modbusRequest ModbusRequest)
}

func AddInfluxTaskForDevice(modbusRequest ModbusRequest) {
	if strings.HasPrefix(modbusRequest.RequestType, "Task") {
		TaskInfluxdb(modbusRequest.RequestType, modbusRequest.Value)
	} else {
		fmt.Println("не относится к influxdb")
	}
}

func TaskInfluxdb(influxType string, requestValue map[string]interface{}) {
	url := "http://influxdb:8086/api/v2/tasks"
	authToken := "Token"
	orgID := "ORGid"

	device := requestValue["name"].(string) // девайс

	sensors := requestValue["sensors"].([]interface{})
	firstSensor := sensors[0].(map[string]interface{})
	nameMetric := firstSensor["name"].(string) // название метрики

	lowThresholdValue := firstSensor["low_threshold_value"].(float64) // нижняя граница
	stat := requestValue["status_task"].(string)                      // статус таски

	fulldevice1d := device + "_1d_" + nameMetric
	fulldevice1h := device + "_1h_" + nameMetric
	text1h := fmt.Sprintf("import \"date\"\nimport \"generate\"\n\noption task = {name: \"%s\", every: 1h, offset: 1m}\n\n//option now = () => 2023-12-18T10:00:00Z\nvBucket = \"Modbus\"\n\nvMeasurement = \"modbus\"\n\nvName = \"%s\"\n\nvField = \"%s\"\n\nvStart = date.sub(d: 1h, from: now())\n\nvStop = now()\n\nvTelegrafTRH = 200000000\n\nvValueTRH = %f\n\nvElapsedTRH = 7000000000\n\ndummy1 = generate.from(count: 2, fn: (n) => n + 1, start: vStart, stop: vStop)\n\ndummy2 =\n    dummy1\n        |> range(start: vStart, stop: vStop)\n        |> map(\n            fn: (r) =>\n                ({r with _time:\n                        if r._value == 2 then\n                            vStop\n                        else\n                            vStart,\n                }),\n        )\n        |> elapsed(unit: 1ns)\n        |> keep(columns: [\"elapsed\"])\n\nt1 =\n    from(bucket: vBucket)\n        |> range(start: vStart, stop: vStop)\n        |> filter(\n            fn: (r) =>\n                r[\"_measurement\"] == vMeasurement and r[\"name\"] == vName and r[\"_field\"] == vField,\n        )\n        |> elapsed(unit: 1ns)\n        |> filter(fn: (r) => r._value >= vValueTRH and r.elapsed < vTelegrafTRH)\n        |> sum(column: \"elapsed\")\n        |> map(fn: (r) => ({r with elapsed: -r.elapsed}))\n        |> keep(columns: [\"elapsed\"])\n\nt2 =\n    from(bucket: vBucket)\n        |> range(start: vStart, stop: vStop)\n        |> filter(\n            fn: (r) =>\n                r[\"_measurement\"] == vMeasurement and r[\"name\"] == vName and r[\"_field\"] == vField,\n        )\n        |> elapsed(unit: 1ns)\n        |> filter(fn: (r) => r._value < vValueTRH and r.elapsed < vTelegrafTRH)\n        |> sum(column: \"elapsed\")\n        |> map(fn: (r) => ({r with elapsed: -r.elapsed}))\n        |> keep(columns: [\"elapsed\"])\n\nt3 =\n    union(tables: [dummy2, t1, t2])\n        |> sum(column: \"elapsed\")\n        |> map(fn: (r) => ({r with _time: vStart}))\n        |> map(\n            fn: (r) =>\n                ({r with elapsed:\n                        if r.elapsed > vElapsedTRH then\n                            r.elapsed\n                        else\n                            0,\n                }),\n        )\n        |> map(fn: (r) => ({r with elapsed: r.elapsed / 1000000000}))\n        |> map(fn: (r) => ({r with _field: \"elapsed_noconn\"}))\n\nt5 =\n    from(bucket: vBucket)\n        |> range(start: vStart, stop: vStop)\n        |> filter(\n            fn: (r) =>\n                r[\"_measurement\"] == vMeasurement and r[\"name\"] == vName and r[\"_field\"] == vField,\n        )\n        |> elapsed(unit: 1ns)\n        |> filter(fn: (r) => r._value > vValueTRH and r.elapsed < vTelegrafTRH)\n        |> sum(column: \"elapsed\")\n        |> map(fn: (r) => ({r with _time: vStart}))\n        |> keep(columns: [\"_time\", \"elapsed\"])\n        |> map(fn: (r) => ({r with elapsed: r.elapsed / 1000000000}))\n        |> map(fn: (r) => ({r with _field: \"elapsed_work\"}))\n\nt6 =\n    from(bucket: vBucket)\n        |> range(start: vStart, stop: vStop)\n        |> filter(\n            fn: (r) =>\n                r[\"_measurement\"] == vMeasurement and r[\"name\"] == vName and r[\"_field\"] == vField,\n        )\n        |> elapsed(unit: 1ns)\n        |> filter(fn: (r) => r._value < vValueTRH and r.elapsed < vTelegrafTRH)\n        |> sum(column: \"elapsed\")\n        |> map(fn: (r) => ({r with _time: vStart}))\n        |> keep(columns: [\"_time\", \"elapsed\"])\n        |> map(fn: (r) => ({r with elapsed: r.elapsed / 1000000000}))\n        |> map(fn: (r) => ({r with _field: \"elapsed_nowork\"}))\n\nunion(tables: [t3, t5, t6])\n    |> rename(columns: {elapsed: \"_value\"})\n    |> map(fn: (r) => ({r with _measurement: \"%s\"}))\n    |> to(\n        bucket: \"AnalyticsV3_1h\",\n        org: \"ORG\",\n        token:\n            \"token\",\n        host: \"http://localhost:8086\",\n    )\n", fulldevice1h, device, nameMetric, lowThresholdValue, fulldevice1h)
	text1d := fmt.Sprintf("import \"date\"\nimport \"timezone\"\n\noption task = {name: \"%s\", every: 1d, offset: 10m}\n\n//option now = () => 2023-12-19T00:00:00.000Z\noption location = timezone.location(name: \"Asia/Yekaterinburg\")\n\nvSubStart = date.sub(d: 5h20m, from: now())\n\nVmeasurement = \"%s\"\nvStart = date.truncate(t: vSubStart, unit: 1d)\nvStop = date.add(d: 24h, to: vStart)\n\nfrom(bucket: \"AnalyticsV3_1h\")\n    |> range(start: vStart, stop: vStop)\n    |> filter(fn: (r) => r[\"_measurement\"] == Vmeasurement and r[\"_field\"] == \"elapsed_work\")\n    |> sum()\n    |> map(fn: (r) => ({r with _value: float(v: r._value) / 3600.0}))\n    |> map(fn: (r) => ({r with _time: vStart}))\n    |> to(\n        bucket: \"AnalyticsV3_1d\",\n        org: \"ORG\",\n        token:\n            \"token\",\n        host: \"http://localhost:8086\",\n    )\n\nfrom(bucket: \"AnalyticsV3_1h\")\n    |> range(start: vStart, stop: vStop)\n    |> filter(fn: (r) => r[\"_measurement\"] == Vmeasurement and r[\"_field\"] == \"elapsed_nowork\")\n    |> sum()\n    |> map(fn: (r) => ({r with _value: float(v: r._value) / 3600.0}))\n    |> map(fn: (r) => ({r with _time: vStart}))\n    |> to(\n        bucket: \"AnalyticsV3_1d\",\n        org: \"ORG\",\n        token:\n            \"token\",\n        host: \"http://localhost:8086\",\n    )\n\nfrom(bucket: \"AnalyticsV3_1h\")\n    |> range(start: vStart, stop: vStop)\n    |> filter(fn: (r) => r[\"_measurement\"] == Vmeasurement and r[\"_field\"] == \"elapsed_noconn\")\n    |> sum()\n    |> map(fn: (r) => ({r with _value: float(v: r._value) / 3600.0}))\n    |> map(fn: (r) => ({r with _time: vStart}))\n    |> to(\n        bucket: \"AnalyticsV3_1d\",\n        org: \"ORG\",\n        token:\n            \"token\",\n        host: \"http://localhost:8086\",\n    )\n", fulldevice1d, fulldevice1h)

	if stat == "active" {
		createTask(fulldevice1d, url, authToken, orgID, text1h, "1d", "10m")
		createTask(fulldevice1h, url, authToken, orgID, text1d, "1h", "1m")
	} else if stat == "delete" {
		deleteTask(fulldevice1d, authToken, url)
		deleteTask(fulldevice1h, authToken, url)
	} else {
		fmt.Println("Неизвестная команда")
	}
}

func createTask(fulldevice string, url string, authToken string, orgID string, text string, t1 string, t2 string) {
	data := map[string]interface{}{
		"orgID":  orgID,
		"name":   fulldevice,
		"every":  t1,
		"offset": t2,
		"flux":   text,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Token "+authToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var taskResp TaskResponse
	err = json.Unmarshal(body, &taskResp)
	if err != nil {
		panic(err)
	}

	//fmt.Println("ID в ответе:", taskResp.ID)
	//fmt.Println("response Body:", string(body))

	err1 := saveDataToFile(fulldevice, taskResp.ID)
	if err != nil {
		fmt.Println("Ошибка сохранения данных в файл:", err1)
	} else {
		fmt.Println("Данные успешно сохранены в файл.")
	}
}

// Ниже функ для удаления (временно переводит таску в статус inactiv) нужны "status_task"
func deleteTask(device string, authToken string, url string) {
	DeviceID, _ := getDeviceID(device)
	urld := url + "/" + DeviceID

	req, err := http.NewRequest("DELETE", urld, nil)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return
	}

	req.Header.Set("Authorization", "Token "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Task deleted successfully")
}

// Ниже функ для изменения (временно переводит таску в статус active) нужны "task_type" и "status_task"
func updateTask(device string) {
	DeviceID, _ := getDeviceID(device)
	//fmt.Println("DeviceID: ", DeviceID)
	url := "http://localhost:8086/api/v2/tasks/" + DeviceID
	authToken := "token"
	payload := []byte(`{"status": "active"}`) // Пример изменения статуса таски

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
}

// Ниже функ для получения id task из txt файла который должен быть прописан в compose
func getDeviceID(device string) (string, error) {
	data, err := os.ReadFile("data.txt")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) == 2 && parts[0] == device {
			return parts[1], nil
		}
	}
	return "", fmt.Errorf("Device ID not found for %s", device)
}

// Ниже функ для сохр id task в txt файл который должен быть прописан в compose
func saveDataToFile(device string, id string) error {
	filePath := "data.txt"
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data := device + ":" + id + "\n"
	_, err = file.WriteString(data)
	if err != nil {
		return err
	}
	return nil
}
