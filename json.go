package utils

import "encoding/json"

// utilJson JSON 编解码工具（标准库 encoding/json 封装）。
type utilJson struct{}

// ToJson 把任意对象序列化为 JSON 字节数组。
// 场景：写入存储前的编码；需要字符串时用 ToJsonString。
func (j *utilJson) ToJson(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

// ToJsonString 把任意对象序列化为紧凑 JSON 字符串，失败返回空串。
// 场景：日志打印参数、拼装简单报文。
func (j *utilJson) ToJsonString(obj interface{}) string {
	data, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(data)
}

// ToPrettyJson 序列化为带缩进的 JSON 字节数组（制表符缩进）。
// 场景：写入配置文件。
func (j *utilJson) ToPrettyJson(obj interface{}) ([]byte, error) {
	return json.MarshalIndent(obj, "", "\t")
}

// ToPrettyJsonString 序列化为带缩进的 JSON 字符串，失败返回空串。
// 场景：调试输出、生成人读得懂的示例报文。
func (j *utilJson) ToPrettyJsonString(obj interface{}) string {
	data, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		return ""
	}
	return string(data)
}

// FromJson 把 JSON 字节数组反序列化到对象 t（传指针）。
// 场景：解析第三方接口响应体。
func (j *utilJson) FromJson(data []byte, obj interface{}) error {
	return json.Unmarshal(data, obj)
}

// Valid 校验 data 是否为合法 JSON。
// 场景：解析前预校验，避免对非法报文调用 FromJson 产生半解析状态。
func (j *utilJson) Valid(data []byte) bool {
	return json.Valid(data)
}
