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
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

/****************************************************
* INSTANCES
****************************************************/
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

func neo4jGetInstance(token string, instance string) (map[string]interface{}, int, error) {
	r, err := neo4jAuraHTTPRequest(token, "GET", "https://api.neo4j.io/v1/instances/"+instance, "")
	if err != nil {
		return nil, r.StatusCode, err
	}
	defer r.Body.Close()
	retMap, err := responseToMap(r)
	if err != nil {
		return nil, r.StatusCode, err
	}
	if _, ok := retMap["errors"]; ok {
		return nil, r.StatusCode, fmt.Errorf("%d - %s", r.StatusCode, retMap["errors"].([]interface{})[0].(map[string]interface{})["reason"])
	} else {
		return retMap, r.StatusCode, err
	}
}

func neo4jDeleteInstance(ctx context.Context, token string, instance string) (map[string]interface{}, error) {
	r, err := neo4jAuraHTTPRequest(token, "DELETE", "https://api.neo4j.io/v1/instances/"+instance, "")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		deleteresp, err := responseToMap(r)
		if err != nil {
			return nil, err
		}
		_, err = neo4jWaitForActionToComplete(ctx, token, instance, "delete", "instance")
		if err != nil {
			return nil, err
		}
		return deleteresp, err
	} else {
		message, err := getResponseErrorMessage(r)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%d - %s", r.StatusCode, message)
	}
}

func neo4jCreateInstance(ctx context.Context, token string, version string, region string, memory string, name string, instancetype string, tenantid string, cloudprovider string, cmk string, vectorOptimized bool, gdsPlugin bool) (map[string]interface{}, error) {
	exists, err := neo4jDoesInstanceExistForTenant(token, tenantid, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("instance %s already exists", name)
	}

	// Construct the payload as a map
	payloadMap := map[string]interface{}{
		"version":        version,
		"region":         region,
		"memory":         memory,
		"name":           name,
		"type":           instancetype,
		"tenant_id":      tenantid,
		"cloud_provider": cloudprovider,
	}

	// Conditionally include `customer_managed_key_id` if it is not empty
	if cmk != "" {
		payloadMap["customer_managed_key_id"] = cmk
	}

	// Conditionally include `vector_optimized` if it is not nil
	if vectorOptimized {
		payloadMap["vector_optimized"] = vectorOptimized
	}

	// Conditionally include `graph_analytics_plugin` if it is not nil
	if gdsPlugin {
		payloadMap["graph_analytics_plugin"] = gdsPlugin
	}

	tflog.Debug(ctx, fmt.Sprintf("create instance payload %v.", payloadMap))
	// Convert the map to a JSON string
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	r, err := neo4jAuraHTTPRequest(token, "POST", "https://api.neo4j.io/v1/instances", string(payloadBytes))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		resp, err := responseToMap(r)
		if err != nil {
			return nil, err
		}
		instance := resp["data"].(map[string]interface{})["id"].(string)
		n4jpwd := resp["data"].(map[string]interface{})["password"]
		resp, err = neo4jWaitForActionToComplete(ctx, token, instance, "create", "instance")
		if err != nil {
			return nil, err
		}
		resp["data"].(map[string]interface{})["n4jpwd"] = n4jpwd
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
	resp, err = neo4jWaitForActionToComplete(ctx, token, instance, action, "instance")
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

func neo4jRenameInstance(token string, instance string, name string) (map[string]interface{}, error) {
	payload := `{"name": "` + name + `"}`
	r, err := neo4jAuraHTTPRequest(token, "PATCH", "https://api.neo4j.io/v1/instances/"+instance, payload)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		resp, err := responseToMap(r)
		return resp, err
	}
	return nil, fmt.Errorf("error renaming instance %d", r.StatusCode)
}

func neo4jUpdate(ctx context.Context, token string, instance string, payload string, description string, hasResp bool) (map[string]interface{}, int, error) {
	tflog.Info(ctx, fmt.Sprintf("neo4jUpdate %s: %s", description, payload))
	r, err := neo4jAuraHTTPRequest(token, "PATCH", "https://api.neo4j.io/v1/instances/"+instance, payload)
	if err != nil {
		return nil, r.StatusCode, err
	}
	defer r.Body.Close()
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		if hasResp {
			resp, err := neo4jWaitForActionToComplete(ctx, token, instance, "update", "instance")
			if err != nil {
				return nil, r.StatusCode, err
			}
			return resp, http.StatusOK, nil
		} else {
			_, err := neo4jWaitForActionToComplete(ctx, token, instance, "update", "instance")
			if err != nil {
				return nil, r.StatusCode, err
			}
			return nil, http.StatusOK, nil
		}
	}
	return nil, r.StatusCode, fmt.Errorf("error updating %s for instance %s, got http %d", description, instance, r.StatusCode)
}

func neo4jUpdateSecondariesCount(ctx context.Context, token string, instance string, secondaries int64) (map[string]interface{}, int, error) {
	payload := `{"secondaries_count": ` + fmt.Sprintf("%d", secondaries) + `}`
	return neo4jUpdate(ctx, token, instance, payload, "secondaries_count", true)
}

func neo4jUpdateMemory(ctx context.Context, token string, instance string, memory string) (map[string]interface{}, int, error) {
	payload := `{"memory": "` + memory + `"}`
	return neo4jUpdate(ctx, token, instance, payload, "memory", true)
}

func neo4jUpdateIncludeGraphPlugin(ctx context.Context, token string, instance string, graph_analytics_plugin bool) (map[string]interface{}, int, error) {
	payload := `{"graph_analytics_plugin": "` + strconv.FormatBool(graph_analytics_plugin) + `"}`
	return neo4jUpdate(ctx, token, instance, payload, "graph_analytics_plugin", false)
}

func neo4jUpdateCDC(ctx context.Context, token string, instance string, cdc_enrichment_mode string) (map[string]interface{}, int, error) {
	payload := `{"cdc_enrichment_mode": "` + cdc_enrichment_mode + `"}`
	return neo4jUpdate(ctx, token, instance, payload, "cdc_enrichment_mode", false)
}

func neo4jUpdateVectorOptimization(ctx context.Context, token string, instance string, vector_optimized bool) (map[string]interface{}, int, error) {
	payload := `{"vector_optimized": ` + strconv.FormatBool(vector_optimized) + `}`
	return neo4jUpdate(ctx, token, instance, payload, "vector_optimized", false)
}

/****************************************************
* CMK
****************************************************/
func neo4jGetCMKs(token string) (map[string]interface{}, error) {
	r, err := neo4jAuraHTTPRequest(token, "GET", "https://api.neo4j.io/v1/customer-managed-keys", "")
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

func neo4jGetCMK(token string, cmkid string) (map[string]interface{}, int, error) {
	r, err := neo4jAuraHTTPRequest(token, "GET", "https://api.neo4j.io/v1/customer-managed-keys/"+cmkid, "")
	if err != nil {
		return nil, r.StatusCode, err
	}
	defer r.Body.Close()
	retMap, err := responseToMap(r)
	if err != nil {
		return nil, r.StatusCode, err
	}
	if _, ok := retMap["errors"]; ok {
		return nil, r.StatusCode, fmt.Errorf("%d - %s", r.StatusCode, retMap["errors"].([]interface{})[0].(map[string]interface{})["reason"])
	} else {
		return retMap, r.StatusCode, err
	}
}

func neo4jDoesCMKExistForTenant(token string, tenant_id string, name string) (bool, error) {
	cmks, err := neo4jGetCMKs(token)
	if err != nil {
		return false, err
	}

	data, ok := cmks["data"].([]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected type for 'data' or 'data' key not found. are you using the correct tenant_id? %s", tenant_id)
	}

	for _, cmk := range data {
		cmkMap, ok := cmk.(map[string]interface{})
		if !ok {
			continue // or handle the unexpected type case
		}
		if name == cmkMap["name"] {
			return true, nil
		}
	}
	return false, nil
}

func neo4jCreateCMK(ctx context.Context, token string, tenantid string, name string, region string, instancetype string, cloudprovider string, keyid string) (map[string]interface{}, error) {
	exists, err := neo4jDoesCMKExistForTenant(token, tenantid, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("cmk %s already exists", name)
	}

	// Construct the payload as a map
	payloadMap := map[string]interface{}{
		"region":         region,
		"name":           name,
		"instance_type":  instancetype,
		"tenant_id":      tenantid,
		"cloud_provider": cloudprovider,
		"key_id":         keyid,
	}

	tflog.Debug(ctx, fmt.Sprintf("create cmk payload %v.", payloadMap))
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	r, err := neo4jAuraHTTPRequest(token, "POST", "https://api.neo4j.io/v1/customer-managed-keys", string(payloadBytes))
	if err != nil {
		_, err := responseToMap(r)
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		resp, err := responseToMap(r)
		cmkId := resp["data"].(map[string]interface{})["id"].(string)
		resp, err = neo4jWaitForActionToComplete(ctx, token, cmkId, "create", "cmk")
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
	return nil, fmt.Errorf("check neo4jCreateCMK implementation. http %d", r.StatusCode)
}

func neo4jDeleteCMK(ctx context.Context, token string, cmkid string) error {
	r, err := neo4jAuraHTTPRequest(token, "DELETE", "https://api.neo4j.io/v1/customer-managed-keys/"+cmkid, "")
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		_, err = neo4jWaitForActionToComplete(ctx, token, cmkid, "delete", "cmk")
		if err != nil {
			return err
		}
		return err
	} else {
		if err != nil {
			return err
		}
		return fmt.Errorf("delete http status: %d", r.StatusCode)
	}
}

/****************************************************
* HELPER METHODS
****************************************************/
func neo4jWaitForActionToComplete(ctx context.Context, token string, objectid string, action string, object string) (map[string]interface{}, error) {
	tflog.Info(ctx, fmt.Sprintf("running neo4jWaitForActionToComplete with action %s on %s %s.", action, object, objectid))
	tries := 0
	sleepSecInterval := 15
	timeoutMin := 30

	initialStatus, secondaryStatus, completed := "", "", false
	for ok := true; ok; {
		status, statusCode := "", 0
		var resp map[string]interface{}
		var err error

		if object == "instance" {
			resp, statusCode, err = neo4jGetInstance(token, objectid)
		} else if object == "cmk" {
			resp, statusCode, err = neo4jGetCMK(token, objectid)
		}

		tflog.Debug(ctx, fmt.Sprintf("got http status: %d", statusCode))
		if err != nil {
			tflog.Debug(ctx, fmt.Sprintf("%s - %s - %d", resp, err, statusCode))
			if action == "delete" && statusCode == 404 {
				return nil, nil // delete completed, cannot lookup status
			}
			return nil, err
		}
		status = resp["data"].(map[string]interface{})["status"].(string)
		completed = false
		if initialStatus != "" && tries >= 2 {
			if action == "pause" {
				completed = status == "paused"
			} else if action == "resume" || action == "create" || action == "update" {
				if object == "cmk" {
					completed = status == "ready"
				} else if object == "instance" {
					completed = status == "running"
				}
			} else if action == "delete" {
				completed = !(status == "deleting" || status == "destroying" || status == "updating" || status == "pending")
			} else {
				tflog.Warn(ctx, "assuming complete with action")
				completed = true
			}
		} else {
			tflog.Info(ctx, fmt.Sprintf("not completing before try/check %d, current status %s, waiting 30 seconds...", tries+1, status))
			time.Sleep(time.Duration(30) * time.Second)
		}
		if initialStatus == "" {
			initialStatus = status
			tflog.Debug(ctx, fmt.Sprintf("Setting initial status: %s", initialStatus))
		} else if initialStatus != status && secondaryStatus == "" {
			tflog.Debug(ctx, fmt.Sprintf("Status changed from %s to %s", initialStatus, status))
			secondaryStatus = status
		} else if secondaryStatus != "" && status != secondaryStatus {
			tflog.Debug(ctx, fmt.Sprintf("Status changed again from %s to %s", secondaryStatus, status))
			secondaryStatus = status
		}
		tflog.Debug(ctx, fmt.Sprintf("get %s response %v.", object, resp))
		tflog.Debug(ctx, fmt.Sprintf("current status: %s, complete: %t", status, completed))
		if completed {
			tflog.Debug(ctx, fmt.Sprintf("action complete, waiting 60 seconds"))
			time.Sleep(time.Duration(60) * time.Second)
			return resp, nil
		}
		tries = tries + 1
		time.Sleep(time.Duration(sleepSecInterval) * time.Second)
		if tries >= (60/sleepSecInterval)*timeoutMin { // 30 minutes
			checkaction := action[:len(action)-1] + "ing"
			return nil, fmt.Errorf("exceeded max number of tries for successfully %s %s %s", checkaction, object, objectid)
		}
	}

	return nil, nil
}

func getResponseErrorMessage(r *http.Response) (string, error) {
	resp, err := responseToMap(r)
	if err != nil {
		return "", err
	}
	instancedata := resp["errors"].([]interface{})
	fmt.Printf("%v", instancedata)
	message := instancedata[0].(map[string]interface{})["message"]
	return message.(string), nil
}

func responseToMap(r *http.Response) (map[string]interface{}, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(bodyBytes))
	decoder.UseNumber() // Use json.Number for numbers
	resp := make(map[string]interface{})
	if err := decoder.Decode(&resp); err != nil {
		return nil, err
	}

	// Recursively convert json.Number or float64 to int64
	convertedResp := convertNumbers(resp)

	// Assert the type to map[string]interface{}
	if resultMap, ok := convertedResp.(map[string]interface{}); ok {
		return resultMap, nil
	}

	return nil, fmt.Errorf("failed to convert response to map[string]interface{}")
}

// Helper function to recursively convert all json.Number and float64 to int64
func convertNumbers(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// Recursively process each key-value pair in the map
		for key, value := range v {
			v[key] = convertNumbers(value)
		}
		return v
	case []interface{}:
		// Recursively process each element in the array
		for i, value := range v {
			v[i] = convertNumbers(value)
		}
		return v
	case json.Number:
		// Convert json.Number to int64
		if intVal, err := v.Int64(); err == nil {
			return intVal
		}
		return v.String() // If conversion fails, return as string
	case float64:
		// Convert float64 to int64 (if safe)
		return int64(v)
	default:
		// Return other types as-is
		return v
	}
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
	attempt_interval := 15 // seconds
	max_attempts := 5
	http_timeout := 30 // seconds
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

func neo4jSizingEstimate(ctx context.Context, token string, node_count int64, relationship_count int64, instance_type string, algorithm_categories []string) (map[string]interface{}, error) {
	tflog.Info(ctx, fmt.Sprintf("running neo4jSizingEstimate with node: %d, relationship: %d, type: %s, categories: %v", node_count, relationship_count, instance_type, algorithm_categories))
	// Step 1: Build the JSON object as a map
	payloadMap := map[string]interface{}{
		"node_count":           node_count,
		"relationship_count":   relationship_count,
		"instance_type":        instance_type,
		"algorithm_categories": algorithm_categories,
	}
	tflog.Debug(ctx, fmt.Sprintf("sizing estimate payload %v.", payloadMap))
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}
	r, err := neo4jAuraHTTPRequest(token, "POST", "https://api.neo4j.io/v1/instances/sizing", string(payloadBytes))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	retMap, err := responseToMap(r)
	if err != nil {
		return nil, err
	}
	if _, ok := retMap["errors"]; ok {
		fmt.Printf("%v", retMap["errors"].([]interface{}))
		return nil, fmt.Errorf("%d - %s", r.StatusCode, retMap["errors"].([]interface{})[0].(map[string]interface{})["reason"])
	} else {
		return retMap, err
	}
}

func neo4jGetProjectConfigurations(token string, tenant_id string) (map[string]interface{}, error) {
	r, err := neo4jAuraHTTPRequest(token, "GET", "https://api.neo4j.io/v1/tenants/"+tenant_id, "")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return responseToMap(r)
}
