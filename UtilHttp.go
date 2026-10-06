package utils

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type HttpResp struct {
	Status int
	Msg    string
	Datas  []byte
}

func (_self *HttpResp) ToString() string {
	if len(_self.Datas) == 0 {
		return ""
	}

	return string(_self.Datas)
}

type utilHttp struct {
	once     sync.Once
	instance *utilHttp
}

func (_self *utilHttp) GetInstance() *utilHttp {
	_self.once.Do(func() {
		_self.instance = new(utilHttp)
	})
	return _self.instance
}

func (_self *utilHttp) SimpleHead(reqUrl string, timeout time.Duration) *HttpResp {

	req, err := http.NewRequest("HEAD", reqUrl, nil)
	if err != nil {
		return &HttpResp{-500, err.Error(), nil}
	}

	client := http.Client{Timeout: timeout}
	_resp, err := client.Do(req)

	if err != nil {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}
	}

	body, err := Json.ToJson(_resp.Header)
	if err == nil {
		return &HttpResp{_resp.StatusCode, "", body}
	} else {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}
	}
}

func (_self *utilHttp) SimpleGet(reqUrl string, timeout time.Duration) *HttpResp {

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return &HttpResp{-500, err.Error(), nil}
	}

	client := http.Client{Timeout: timeout}
	_resp, err := client.Do(req)

	if err != nil {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}
	}

	defer _resp.Body.Close()

	body, err := io.ReadAll(_resp.Body)
	if err == nil {
		return &HttpResp{_resp.StatusCode, "", body}
	} else {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}
	}
}

func (_self *utilHttp) Get(reqUrl string, headerParam map[string]string, timeout time.Duration) (resp *HttpResp, err error) {

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return &HttpResp{-500, err.Error(), nil}, err
	}

	for k, v := range headerParam {
		req.Header.Set(k, v)
	}

	client := http.Client{Timeout: timeout}
	_resp, err := client.Do(req)

	if err != nil {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}, err
	}

	defer _resp.Body.Close()

	body, err := io.ReadAll(_resp.Body)
	if err == nil {
		return &HttpResp{_resp.StatusCode, "", body}, nil
	} else {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}, err
	}

}

func (_self *utilHttp) SimplePost(reqUrl string, timeout time.Duration) *HttpResp {
	req, err := http.NewRequest("POST", reqUrl, nil)
	if err != nil {
		return &HttpResp{-500, err.Error(), nil}
	}

	//req.Header.Set("Content-Type", contentType)

	client := http.Client{Timeout: timeout}
	_resp, err := client.Do(req)

	if err != nil {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}
	}

	defer _resp.Body.Close()

	body, err := io.ReadAll(_resp.Body)
	if err == nil {
		return &HttpResp{_resp.StatusCode, "", body}
	} else {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}
	}
}

func (_self *utilHttp) Post(reqUrl string, headerParam map[string]string, param map[string]string, timeout time.Duration) (resp *HttpResp, err error) {

	var paramBuf bytes.Buffer
	paramBuf.WriteString("curTime=" + Date.GetCurrentTime())
	for k, v := range param {
		paramBuf.WriteString("&" + k + "=" + v)
	}

	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(paramBuf.String()))
	if err != nil {
		return &HttpResp{-500, err.Error(), nil}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headerParam {
		req.Header.Set(k, v)
	}

	client := http.Client{Timeout: timeout}
	_resp, err := client.Do(req)

	if err != nil {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}, err
	}

	defer _resp.Body.Close()

	body, err := io.ReadAll(_resp.Body)
	if err == nil {
		return &HttpResp{_resp.StatusCode, "", body}, nil
	} else {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}, err
	}
}

func (_self *utilHttp) HttpPostJson(reqUrl string, json string, timeout time.Duration) (resp *HttpResp, err error) {
	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(json))
	if err != nil {
		return &HttpResp{-500, err.Error(), nil}, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: timeout}
	_resp, err := client.Do(req)

	if err != nil {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}, err
	}

	defer _resp.Body.Close()

	body, err := io.ReadAll(_resp.Body)
	if err == nil {
		return &HttpResp{_resp.StatusCode, "", body}, nil
	} else {
		return &HttpResp{_resp.StatusCode, err.Error(), nil}, err
	}

}

func (_self *utilHttp) UrlEncode(input string) string {
	if String.IsEmpty(input) {
		return ""
	}
	return url.QueryEscape(input)
}

func (_self *utilHttp) UrlDecode(input string) string {
	if String.IsEmpty(input) {
		return ""
	}
	result, err := url.QueryUnescape(input)
	if err != nil {
		return input
	} else {
		return result
	}
}
