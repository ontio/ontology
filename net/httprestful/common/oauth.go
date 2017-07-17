package common

import (
	. "DNA/common/config"
	"DNA/common/log"
	Err "DNA/net/httprestful/error"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"time"
)

var oauthClient = NewOauthClient()

//config
func GetOauthServerUrl(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	resp["Result"] = Parameters.OauthServerUrl
	return resp
}
func SetOauthServerUrl(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)

	addr, ok := cmd["Url"].(string)
	if !ok {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	if len(addr) > 0 {
		var reg *regexp.Regexp
		pattern := `((http|https)://)(([a-zA-Z0-9\._-]+\.[a-zA-Z]{2,6})|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}))(:[0-9]{1,4})*(/[a-zA-Z0-9\&%_\./-~-]*)?`
		reg = regexp.MustCompile(pattern)
		if !reg.Match([]byte(addr)) {
			resp["Error"] = Err.INVALID_PARAMS
			return resp
		}
	}
	Parameters.OauthServerUrl = addr
	resp["Result"] = Parameters.OauthServerUrl
	return resp
}

func NewOauthClient() *http.Client {
	c := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, time.Second*10)
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(time.Second * 10))
				return conn, nil
			},
			DisableKeepAlives: false,
		},
	}
	return c
}

func OauthRequest(method string, cmd map[string]interface{}, url string) (map[string]interface{}, error) {

	var repMsg = make(map[string]interface{})
	var response *http.Response
	var err error
	switch method {
	case "GET":

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return repMsg, err
		}
		response, err = oauthClient.Do(req)

	case "POST":
		data, err := json.Marshal(cmd)
		if err != nil {
			return repMsg, err
		}
		reqData := bytes.NewBuffer(data)
		req, err := http.NewRequest("POST", url, reqData)
		if err != nil {
			return repMsg, err
		}
		req.Header.Set("Content-type", "application/json")
		response, err = oauthClient.Do(req)
	default:
		return repMsg, err
	}
	if response != nil {
		defer response.Body.Close()

		body, _ := ioutil.ReadAll(response.Body)
		if err := json.Unmarshal(body, &repMsg); err == nil {
			return repMsg, err
		}
	}
	if err != nil {
		return repMsg, err
	}

	return repMsg, err
}
func CheckAccessToken(auth_type, access_token string) (cakey string, errCode int64, result interface{}) {

	if len(Parameters.OauthServerUrl) == 0 {
		return "", Err.SUCCESS, ""
	}
	req := make(map[string]interface{})
	req["token"] = access_token
	req["auth_type"] = auth_type
	rep, err := OauthRequest("GET", req, Parameters.OauthServerUrl+"/"+access_token+"?auth_type="+auth_type)
	if err != nil {
		log.Error("Oauth timeout:", err)
		return "", Err.OAUTH_TIMEOUT, rep
	}
	if errcode, ok := rep["Error"].(float64); ok && errcode == 0 {
		result, ok := rep["Result"].(map[string]interface{})
		if !ok {
			return "", Err.INVALID_TOKEN, rep
		}
		if CAkey, ok := result["CaKey"].(string); ok {
			return CAkey, Err.SUCCESS, rep
		}
	}
	return "", Err.INVALID_TOKEN, rep
}
