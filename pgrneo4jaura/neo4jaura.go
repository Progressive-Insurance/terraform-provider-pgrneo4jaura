/******************************************
* Neo4j Administrative API
* https://neo4j.com/docs/aura/platform/api/specification/
******************************************/
package pgrneo4jaura

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// default return true to be safe. caller should always check for err
func neo4jDoesInstanceExistForTenant(token string, tenant_id string, name string) (bool, error) {
	instances, err := neo4jGetInstances(token, tenant_id)
	if err != nil {
		return true, err
	}
	for _, instance := range instances["data"].([]interface{}) {
		if name == instance.(map[string]interface{})["name"] {
			return true, nil
		}
	}
	return false, nil
}

func neo4jGetInstances(token string, tenant_id string) (map[string]interface{}, error) {
	r, err := neo4jAuraHTTPRequest(token, "GET", "https://api.neo4j.io/v1/instances?tenantId="+tenant_id, "")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return responseToMap(r)
}

func neo4jGetInstance(token string, instance string) (map[string]interface{}, error) {
	r, err := neo4jAuraHTTPRequest(token, "GET", "https://api.neo4j.io/v1/instances/"+instance, "")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	retMap, err := responseToMap(r)
	if err != nil {
		return nil, err
	}
	if _, ok := retMap["errors"]; ok {
		return nil, fmt.Errorf("%d - %s", r.StatusCode, retMap["errors"].([]interface{})[0].(map[string]interface{})["reason"])
	} else {
		return retMap, err
	}
}

func neo4jDeleteInstance(token string, instance string) (map[string]interface{}, error) {
	r, err := neo4jAuraHTTPRequest(token, "DELETE", "https://api.neo4j.io/v1/instances/"+instance, "")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return responseToMap(r)
	} else {
		message, err := getResponseErrorMessage(r)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%d - %s", r.StatusCode, message)
	}
}

func neo4jCreateInstance(token string, version string, region string, memory string, name string, instancetype string, tenantid string, cloudprovider string) (map[string]interface{}, error) {
	exists, err := neo4jDoesInstanceExistForTenant(token, tenantid, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("instance %s already exists", name)
	}
	payload := `{"version": "` + version + `", "region": "` + region + `", "memory": "` + memory + `", "name": "` + name + `", "type": "` + instancetype + `", "tenant_id": "` + tenantid + `", "cloud_provider": "` + cloudprovider + `"}`
	r, err := neo4jAuraHTTPRequest(token, "POST", "https://api.neo4j.io/v1/instances", payload)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		resp, err := responseToMap(r)
		if err != nil {
			return nil, err
		}
		// WAIT FOR INSTANCE TO FINISH CREATING
		tries := 0
		for ok := true; ok; {
			instance, err := neo4jGetInstance(token, resp["data"].(map[string]interface{})["id"].(string))
			if err != nil {
				return nil, err
			}
			tries = tries + 1
			status := instance["data"].(map[string]interface{})["status"].(string)
			if status != "running" {
				// if status == "failed" {
				// 	return fmt.Errorf("TODO some failed message. cant reproduce failed status")
				// }
				time.Sleep(5 * time.Second)
				if tries >= 240 {
					return nil, fmt.Errorf("exceeded max amount of time waiting for neo4j aura instance to finish creating/building")
				}
			} else {
				return instance, nil
			}
		}
		// END WAIT FOR INSTANCE TO FINISH CREATING
	} else if r.StatusCode == http.StatusConflict || r.StatusCode == http.StatusBadRequest {
		message, err := getResponseErrorMessage(r)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%s", message)
	}
	return nil, fmt.Errorf("check neo4jCreateInstance implementation. http %d", r.StatusCode)
}

func neo4jResumeInstance(token string, instance string, wait bool) (map[string]interface{}, error) {
	if wait {
		return neo4jInstanceActionWithWait(token, instance, "resume")
	} else {
		resp, _, err := neo4jInstanceAction(token, instance, "resume")
		return resp, err
	}
}

func neo4jPauseInstance(token string, instance string, wait bool) (map[string]interface{}, error) {
	if wait {
		return neo4jInstanceActionWithWait(token, instance, "pause")
	} else {
		resp, _, err := neo4jInstanceAction(token, instance, "pause")
		return resp, err
	}
}

func neo4jInstanceActionWithWait(token string, instance string, action string) (map[string]interface{}, error) {
	if action != "pause" && action != "resume" {
		return nil, nil
	}
	resp, ismsgerr, err := neo4jInstanceAction(token, instance, action)
	tries := 0
	sleepSecInterval := 5
	timeoutMin := 30
	for ok := true; ok; {
		if err != nil && !ismsgerr {
			return nil, err
		} else if err != nil && ismsgerr {
			if strings.Contains(err.Error(), "is not running") && action == "pause" {
				return resp, nil
			}
			if strings.Contains(err.Error(), "is not paused") && action == "resume" {
				return resp, nil
			}
		}
		checkaction := action[:len(action)-1] + "ing"
		takingAction := true
		if action == "pause" {
			takingAction = resp["data"].(map[string]interface{})["status"] == "pausing"
		} else {
			takingAction = resp["data"].(map[string]interface{})["status"] == "resuming"
		}
		if takingAction {
			tries = tries + 1
			time.Sleep(time.Duration(sleepSecInterval) * time.Second)
			if tries >= (60/sleepSecInterval)*timeoutMin { // 20 minutes
				return nil, fmt.Errorf("exceeded max number of tries for successfully %s instance %s", checkaction, instance)
			}
			resp, ismsgerr, err = neo4jInstanceAction(token, instance, action)
		}
	}
	return resp, err
}

/*
return values:
 1. response from last API call
 2. whether the 3rd return value is a "known" error. returning true can have the caller
    loop to periodically check the instance status message
 3. an error will be artificially created if calling pause while already pausing or resume while already resuming
    so that callers of this method can choose to wait for status completion or not. if #2 is true then checking
    the error message will indiciate the instance status. if #2 is false and an error is returned then the error
    is unexpected.
*/
func neo4jInstanceAction(token string, instance string, action string) (map[string]interface{}, bool, error) {
	r, err := neo4jAuraHTTPRequest(token, "POST", "https://api.neo4j.io/v1/instances/"+instance+"/"+action, "")
	if err != nil {
		return nil, false, err
	}
	defer r.Body.Close()
	if r.StatusCode == http.StatusOK || r.StatusCode == http.StatusAccepted { // have seen both during pause/resume
		resp, err := responseToMap(r)
		return resp, false, err
	} else if r.StatusCode == http.StatusConflict || r.StatusCode == http.StatusBadRequest {
		resp, err := responseToMap(r)
		if err != nil {
			return nil, false, err
		}
		instancedata := resp["errors"].([]interface{})
		checkaction := action[:len(action)-1] + "ing"
		message := instancedata[0].(map[string]interface{})["message"]
		instance, err := neo4jGetInstance(token, instance)
		if err != nil {
			return nil, false, err
		}
		if message == ("The database is currently undergoing an operation: " + checkaction) {
			return instance, false, err
		} else {
			return instance, true, fmt.Errorf("%s", message)
		}
	}
	return nil, false, fmt.Errorf("check neo4jInstanceAction implementation. http %d", r.StatusCode)
}

func getResponseErrorMessage(r *http.Response) (string, error) {
	resp, err := responseToMap(r)
	if err != nil {
		return "", err
	}
	instancedata := resp["errors"].([]interface{})
	message := instancedata[0].(map[string]interface{})["message"]
	return message.(string), nil
}

func responseToMap(r *http.Response) (map[string]interface{}, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	resp := make(map[string]interface{})
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func getNeo4jAuraAuthToken(client_id string, client_secret string) (string, error) {
	r, err := neo4jAuraHTTPRequestWithBasicAuth(client_id, client_secret, "POST", "https://api.neo4j.io/oauth/token", `{"grant_type":"client_credentials"}`)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	resp, err := responseToMap(r)
	if err != nil {
		return "", err
	}
	if r.StatusCode == http.StatusOK {
		return resp["access_token"].(string), nil
	}
	strresp, err := mapToString(resp)
	if err != nil {
		return "", err
	}
	return "", fmt.Errorf("error during authentication %d : %s - client id: %s", r.StatusCode, strresp, client_id)
}

func neo4jAuraHTTPRequestWithBasicAuth(client_id string, client_secret string, method string, url string, messagebody string) (*http.Response, error) {
	return neo4jAuraHTTPRequestHelper("", method, url, messagebody, client_id, client_secret)
}

func neo4jAuraHTTPRequest(token string, method string, url string, messagebody string) (*http.Response, error) {
	return neo4jAuraHTTPRequestHelper(token, method, url, messagebody, "", "")
}

func neo4jAuraHTTPRequestHelper(token string, method string, url string, messagebody string, client_id string, client_secret string) (*http.Response, error) {
	attempt := 0
	attempt_interval := 5 // seconds
	max_attempts := 5
	http_timeout := 10 // seconds
	for ok := true; ok; {
		attempt = attempt + 1

		client := &http.Client{Timeout: time.Duration(http_timeout) * time.Second}
		req, err := http.NewRequest(method,
			url,
			bytes.NewBuffer([]byte(messagebody)))
		if err != nil {
			return &http.Response{}, err
		}
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		} else {
			req.SetBasicAuth(client_id, client_secret)
		}
		r, err := client.Do(req)
		if err != nil {
			if err == context.DeadlineExceeded {
				if attempt >= max_attempts {
					return nil, fmt.Errorf("unable to execute %s request after %d attempts\n\turl: %s\n\tpayload: %s", method, max_attempts, url, messagebody)
				}
				time.Sleep(time.Duration(attempt_interval) * time.Second)
			} else {
				return &http.Response{}, err
			}
		} else {
			return r, nil
		}
	}
	return nil, fmt.Errorf("unexpected error")
}

func mapToString(inputMap map[string]interface{}) (string, error) {
	jsonBytes, err := json.Marshal(inputMap)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func neo4jRenameInstance(token string, instance string, name string) (map[string]interface{}, error) {
	payload := `{"name": "` + name + `"}`
	r, err := neo4jAuraHTTPRequest(token, "PATCH", "https://api.neo4j.io/v1/instances/"+instance, payload)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode == http.StatusOK || r.StatusCode == http.StatusAccepted { // have seen both during pause/resume
		resp, err := responseToMap(r)
		return resp, err
	}
	return nil, fmt.Errorf("error renaming instance %d", r.StatusCode)
}
