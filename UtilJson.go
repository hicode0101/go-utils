package utils

import "encoding/json"

type utilJson struct {
}

func (_self *utilJson) ToJson(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

func (_self *utilJson) ToPrettyJson(obj interface{}) ([]byte, error) {
	return json.MarshalIndent(obj, "", "	")
}

func (_self *utilJson) ToPrettyJsonString(obj interface{}) string {
	jsonBytes, err := json.MarshalIndent(obj, "", "	")
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

func (_self *utilJson) FromJson(data []byte, t interface{}) error {
	return json.Unmarshal(data, t)
}
