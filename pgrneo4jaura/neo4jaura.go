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
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// default return true to be safe. caller should always check for err
func neo4jDoesInstanceExistForTenant(token string, tenant_id string, name string) (bool, error) {
	instances, err := neo4jGetInstances(token, tenant_id)
	if err != nil {
		return false, err
	}

	// Check if the "data" key exists and is of the expected type
	data, ok := instances["data"].([]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected type for 'data' or 'data' key not found. are you using the correct tenant_id? %s", tenant_id)
	}

	for _, instance := range data {
		instanceMap, ok := instance.(map[string]interface{})
		if !ok {
			continue // or handle the unexpected type case
		}
		if name == instanceMap["name"] {
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

func neo4jCreateInstance(ctx context.Context, token string, version string, region string, memory string, name string, instancetype string, tenantid string, cloudprovider string) (map[string]interface{}, error) {
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
		instance := resp["data"].(map[string]interface{})["id"].(string)
		resp, err = neo4jWaitForActionToComplete(ctx, token, instance, "create")
		if err != nil {
			return nil, err
		}
		return resp, err
	} else if r.StatusCode == http.StatusConflict || r.StatusCode == http.StatusBadRequest {
		message, err := getResponseErrorMessage(r)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%s", message)
	}
	return nil, fmt.Errorf("check neo4jCreateInstance implementation. http %d", r.StatusCode)
}

func neo4jResumeInstance(ctx context.Context, token string, instance string, wait bool) (map[string]interface{}, error) {
	if wait {
		return neo4jInstanceActionWithWait(ctx, token, instance, "resume")
	} else {
		return neo4jInstanceAction(token, instance, "resume")
	}
}

func neo4jPauseInstance(ctx context.Context, token string, instance string, wait bool) (map[string]interface{}, error) {
	if wait {
		return neo4jInstanceActionWithWait(ctx, token, instance, "pause")
	} else {
		return neo4jInstanceAction(token, instance, "pause")
	}
}

func neo4jInstanceActionWithWait(ctx context.Context, token string, instance string, action string) (map[string]interface{}, error) {
	tflog.Info(ctx, fmt.Sprintf("running neo4jInstanceActionWithWait with action %s on instance %s.", action, instance))
	if action != "pause" && action != "resume" {
		return nil, nil
	}
	resp, err := neo4jInstanceAction(token, instance, action)
	if err != nil {
		return nil, err
	}
	resp, err = neo4jWaitForActionToComplete(ctx, token, instance, action)
	return resp, err
}

func neo4jWaitForActionToComplete(ctx context.Context, token string, instance string, action string) (map[string]interface{}, error) {
	tries := 0
	sleepSecInterval := 5
	timeoutMin := 30

	for ok := true; ok; {
		resp, err := neo4jGetInstance(token, instance)
		if err != nil {
			return nil, err
		}
		tflog.Debug(ctx, fmt.Sprintf("get instance response %v.", resp))
		status := resp["data"].(map[string]interface{})["status"].(string)
		checkaction := action[:len(action)-1] + "ing"
		takingAction := false
		if action == "pause" {
			takingAction = status == "pausing"
		} else if action == "resume" {
			takingAction = status == "resuming" || status == "restoring"
		} else if action == "create" {
			takingAction = status == "creating"
		} else if action == "update" {
			takingAction = status == "updating" || status == "resizing"
		} else if action == "run" {
			takingAction = status == "running"
		}
		if !takingAction {
			return resp, err
		}
		tries = tries + 1
		time.Sleep(time.Duration(sleepSecInterval) * time.Second)
		if tries >= (60/sleepSecInterval)*timeoutMin { // 20 minutes
			return nil, fmt.Errorf("exceeded max number of tries for successfully %s instance %s", checkaction, instance)
		}
	}

	return nil, nil
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
func neo4jInstanceAction(token string, instance string, action string) (map[string]interface{}, error) {
	r, err := neo4jAuraHTTPRequest(token, "POST", "https://api.neo4j.io/v1/instances/"+instance+"/"+action, "")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode == http.StatusOK || r.StatusCode == http.StatusAccepted { // have seen both during pause/resume
		resp, err := responseToMap(r)
		return resp, err
	} else if r.StatusCode == http.StatusConflict || r.StatusCode == http.StatusBadRequest {
		message, err := getResponseErrorMessage(r)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%s", message)
	}
	return nil, fmt.Errorf("check neo4jInstanceAction implementation. http %d", r.StatusCode)
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

func neo4jUpdateMemory(ctx context.Context, token string, instance string, memory string) (map[string]interface{}, error) {
	payload := `{"memory": "` + memory + `"}`
	r, err := neo4jAuraHTTPRequest(token, "PATCH", "https://api.neo4j.io/v1/instances/"+instance, payload)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode == http.StatusOK || r.StatusCode == http.StatusAccepted {
		_, err := neo4jWaitForActionToComplete(ctx, token, instance, "run") //wait for instance status to change from running to updating
		if err != nil {
			return nil, err
		}
		return neo4jWaitForActionToComplete(ctx, token, instance, "update") //wait for updating and resizing status to transition to running
	}
	return nil, fmt.Errorf("error updating memory for instance %s, got http %d", instance, r.StatusCode)
}
