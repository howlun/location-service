package location

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/iknowhtml/locationtracker/pkg/common"
)

type LocationService struct {
	client *LocationClient
}

func (ls *LocationService) Init(c *LocationClient) error {
	if c != nil {
		log.Printf("On connection: %v\n", c)
		ls.client = c
	} else {
		ls.client = new(LocationClient)
		err := ls.client.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ls *LocationService) GetObject(
	key Object_Collection,
	id int32) (*GetObjectResponseObject, error) {

	var respObj GetObjectResponseObject
	conn := ls.client.pool.Get()
	defer conn.Close()

	if key == "" {
		return nil, errors.New("Key is empty")
	}
	if id == 0 {
		return nil, errors.New("Id is not set")
	}

	// generate command args
	objID := GenerateLocationObjectId("", id)
	commandType := "GET"
	commandArgs := []interface{}{key, objID, "WITHFIELDS"}

	// send command to redis to update cache
	log.Printf("Cmd: %s %v\n", commandType, commandArgs)
	res, err := redis.Bytes(conn.Do(commandType, commandArgs...))
	if err != nil {
		return nil, err
	}
	log.Printf("Get Object successful - key: %s, id: %s\n", key, objID)
	//log.Println(res)
	// decode response to json object of ley value pair (string, generic)
	err = json.Unmarshal(res, &respObj)
	if err != nil {
		return nil, err
	}

	log.Println(string(res))
	log.Println(respObj)

	// Return the struct object of result
	respObj.ObjectCollection = key
	return &respObj, nil
}

func (ls *LocationService) SetField(
	key Object_Collection,
	id int32,
	fields LocationObject_Fields) (*SetFieldResponseObject, error) {

	var respObj SetFieldResponseObject
	conn := ls.client.pool.Get()
	defer conn.Close()

	if key == "" {
		return nil, errors.New("Key is empty")
	}
	if id == 0 {
		return nil, errors.New("Id is not set")
	}
	if fields == nil {
		return nil, errors.New("Field object is nil")
	}

	// generate command args
	objID := GenerateLocationObjectId("", id)
	commandType := "FSET"
	commandArgs := []interface{}{key, objID}
	for k, v := range fields {
		commandArgs = append(commandArgs, k)
		commandArgs = append(commandArgs, v)
	}
	// send command to redis to update cache
	log.Printf("Cmd: %s %v\n", commandType, commandArgs)
	res, err := redis.Bytes(conn.Do(commandType, commandArgs...))
	if err != nil {
		return nil, err
	}
	log.Printf("Set Field successful - key: %s, id: %s\n", key, objID)
	//log.Println(res)
	// decode response to json object of ley value pair (string, generic)
	err = json.Unmarshal(res, &respObj)
	if err != nil {
		return nil, err
	}

	log.Println(string(res))
	log.Println(respObj)
	// Return value is the integer count of how many fields actually changed their values
	return &respObj, nil
}

func (ls *LocationService) SetObject(
	key Object_Collection,
	id int32,
	obj *LocationObject,
	fields LocationObject_Fields) (*SetObjectResponseObject, error) {

	var respObj SetObjectResponseObject
	conn := ls.client.pool.Get()
	defer conn.Close()

	if key == "" {
		return nil, errors.New("Key is empty")
	}
	if id == 0 {
		return nil, errors.New("Id is not set")
	}
	if obj == nil {
		return nil, errors.New("Location object is nil")
	}

	// convert LocationObject to JSON
	jsonObj, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	log.Printf("JSON object created: %s\n", string(jsonObj))

	// send command to redis to update cache
	objID := GenerateLocationObjectId("", id)
	commandType := "SET"
	commandArgs := []interface{}{key, objID}
	if fields != nil && len(fields) > 0 {
		for k, v := range fields {
			commandArgs = append(commandArgs, "FIELD")
			commandArgs = append(commandArgs, k)
			commandArgs = append(commandArgs, v)
		}
	}
	// Pass by POINT
	commandArgs = append(commandArgs, "POINT")
	commandArgs = append(commandArgs, obj.Coordinates[0]) // lat
	commandArgs = append(commandArgs, obj.Coordinates[1]) // lng

	log.Printf("Cmd: %s %v\n", commandType, commandArgs)
	res, err := redis.Bytes(conn.Do(commandType, commandArgs...))
	if err != nil {
		return nil, err
	}
	log.Println(string(res))
	log.Printf("Set Object successful - key: %s, id: %s\n", key, objID)

	// decode response to json object of ley value pair (string, generic)
	err = json.Unmarshal(res, &respObj)
	if err != nil {
		return nil, err
	}

	log.Println(respObj)

	// return string 'OK'
	return &respObj, nil
}

func (ls *LocationService) NearbyObject(
	key Object_Collection,
	point_lat float32,
	point_lng float32,
	radius int32,
	limit int32,
	whereList []WhereConditionFieldObject,
	whereInList []WhereInConditionFieldObject) (*NearbyObjectResponseObject, error) {

	if key == "" {
		return nil, errors.New("Key is empty")
	}
	if radius <= 0 {
		return nil, errors.New("Search radius is not set")
	}
	if limit <= 0 {
		return nil, errors.New("Search limit is not set")
	}

	var respObj NearbyObjectResponseObject
	conn := ls.client.pool.Get()
	defer conn.Close()

	// send command to redis to update cache
	commandType := "NEARBY"
	commandArgs := []interface{}{key}
	if limit > 0 {
		commandArgs = append(commandArgs, "LIMIT")
		commandArgs = append(commandArgs, limit)
	}

	for _, c := range whereList {
		commandArgs = append(commandArgs, "WHERE")
		commandArgs = append(commandArgs, c.FieldName)
		commandArgs = append(commandArgs, c.Min)
		commandArgs = append(commandArgs, c.Max)
	}

	for _, wi := range whereInList {
		commandArgs = append(commandArgs, "WHEREIN")
		if count := len(wi.Values); count > 0 {
			commandArgs = append(commandArgs, wi.FieldName)
			commandArgs = append(commandArgs, count)
			for _, wiv := range wi.Values {
				commandArgs = append(commandArgs, wiv)
			}
		}
	}

	commandArgs = append(commandArgs, "POINT")
	commandArgs = append(commandArgs, point_lat)
	commandArgs = append(commandArgs, point_lng)
	commandArgs = append(commandArgs, radius)

	log.Printf("Cmd: %s %v\n", commandType, commandArgs)
	res, err := redis.Bytes(conn.Do(commandType, commandArgs...))
	if err != nil {
		return nil, err
	}

	log.Println(string(res))
	log.Printf("Search Nearby successful - key: %s, lat: %f, lng: %f, radius: %d\n", key, point_lat, point_lng, radius)

	// decode response to json object of ley value pair (string, generic)
	err = json.Unmarshal(res, &respObj)
	if err != nil {
		return nil, err
	}

	log.Println(respObj)

	respObj.ObjectCollection = key
	return &respObj, nil
}

func (ls *LocationService) SetHookSearchFence(
	endPoints []string,
	topicName string,
	searchType string,
	key Object_Collection,
	id int32,
	point_lat float32,
	point_lng float32,
	radius int32,
	detectList map[string]string,
	commandList map[string]string,
	whereList []WhereConditionFieldObject,
	whereInList []WhereInConditionFieldObject) (*HookFenceResponseObject, error) {

	if topicName == "" {
		return nil, errors.New("Hook Name is empty")
	}
	if endPoints == nil || len(endPoints) == 0 {
		return nil, errors.New("Hook Endpoint is empty")
	}
	if searchType == "" {
		return nil, errors.New("Search Type is empty")
	}
	if key == "" {
		return nil, errors.New("Key is empty")
	}
	if radius <= 0 {
		return nil, errors.New("Search radius is not set")
	}

	// send command to redis to update cache
	var respObj HookFenceResponseObject
	objID := GenerateLocationObjectId("", id)
	hookName := GenerateHookName(topicName, string(key), objID)
	conn := ls.client.pool.Get()
	defer conn.Close()

	commandType := "SETHOOK"
	commandArgs := []interface{}{hookName}

	epStr := ""
	for i, ep := range endPoints {
		if i == 0 {
			epStr = common.Concate(epStr, ep, "/", topicName)
		} else {
			epStr = common.Concate(epStr, ",", ep, "/", topicName)
		}
	}
	commandArgs = append(commandArgs, epStr)

	commandArgs = append(commandArgs, searchType)
	commandArgs = append(commandArgs, key)
	commandArgs = append(commandArgs, "MATCH")
	commandArgs = append(commandArgs, objID)

	for _, c := range whereList {
		commandArgs = append(commandArgs, "WHERE")
		commandArgs = append(commandArgs, c.FieldName)
		commandArgs = append(commandArgs, c.Min)
		commandArgs = append(commandArgs, c.Max)
	}

	for _, wi := range whereInList {
		commandArgs = append(commandArgs, "WHEREIN")
		if count := len(wi.Values); count > 0 {
			commandArgs = append(commandArgs, wi.FieldName)
			commandArgs = append(commandArgs, count)
			for _, wiv := range wi.Values {
				commandArgs = append(commandArgs, wiv)
			}
		}
	}

	commandArgs = append(commandArgs, "FENCE")

	detectLen := len(detectList)
	if detectList != nil && detectLen > 0 {
		count := 0
		detectBuffer := new(bytes.Buffer)
		for _, d := range detectList {
			if count == 0 {
				fmt.Fprintf(detectBuffer, "%s", d)
			} else {
				fmt.Fprintf(detectBuffer, ",%s", d)
			}
			count++
		}
		commandArgs = append(commandArgs, "DETECT")
		commandArgs = append(commandArgs, detectBuffer.String())
	}

	commandLen := len(commandList)
	if commandList != nil && commandLen > 0 {
		count := 0
		commandBuffer := new(bytes.Buffer)
		for _, d := range commandList {
			if count == 0 {
				fmt.Fprintf(commandBuffer, "%s", d)
			} else {
				fmt.Fprintf(commandBuffer, ",%s", d)
			}
			count++
		}
		commandArgs = append(commandArgs, "COMMANDS")
		commandArgs = append(commandArgs, commandBuffer.String())
	}

	commandArgs = append(commandArgs, "POINT")
	commandArgs = append(commandArgs, point_lat)
	commandArgs = append(commandArgs, point_lng)
	commandArgs = append(commandArgs, radius)

	log.Printf("Cmd: %s %v\n", commandType, commandArgs)

	res, err := redis.Bytes(conn.Do(commandType, commandArgs...))
	if err != nil {
		return nil, err
	}

	log.Println(string(res))
	log.Printf("Set Hook Search Fence successful - key: %s, id: %s\n", key, objID)

	// decode response to json object of ley value pair (string, generic)
	err = json.Unmarshal(res, &respObj)
	if err != nil {
		return nil, err
	}

	log.Println(respObj)

	return &respObj, nil
}

func (ls *LocationService) DelHookSearchFence(
	topicName string,
	searchType string,
	key Object_Collection,
	id int32) (*HookFenceResponseObject, error) {

	if topicName == "" {
		return nil, errors.New("Hook Name is empty")
	}
	if searchType == "" {
		return nil, errors.New("Search Type is empty")
	}
	if key == "" {
		return nil, errors.New("Key is empty")
	}

	// send command to redis to update cache
	var respObj HookFenceResponseObject
	objID := GenerateLocationObjectId("", id)
	hookName := GenerateHookName(topicName, string(key), objID)
	conn := ls.client.pool.Get()
	defer conn.Close()

	commandType := "DELHOOK"
	commandArgs := []interface{}{hookName}

	log.Printf("Cmd: %s %v\n", commandType, commandArgs)

	res, err := redis.Bytes(conn.Do(commandType, commandArgs...))
	if err != nil {
		return nil, err
	}

	log.Println(string(res))
	log.Printf("Delete Hook Search Fence successful - key: %s, id: %s\n", key, objID)

	// decode response to json object of ley value pair (string, generic)
	err = json.Unmarshal(res, &respObj)
	if err != nil {
		return nil, err
	}

	log.Println(respObj)

	return &respObj, nil
}

func GenerateLocationObjectId(objType string, id int32) string {
	if id == 0 {
		return "*"
	} else {
		if objType != "" {
			return common.String(id) + ":" + objType
		} else {
			return common.String(id)
		}

	}
}

func GenerateHookName(topic string, key string, objId string) string {
	return topic + "_" + key + "_" + objId
}
