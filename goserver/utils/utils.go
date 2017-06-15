package utils

import (
	"encoding/base64"
	"os"
	"fmt"
)

const SERVERDATA = "./serverdata"
const KERNELADIUTOR = SERVERDATA + "/kerneladiutor"
const MANDY = SERVERDATA + "/mandy"

func StringEmpty(text string) bool {
	return len(text) == 0
}

func FileExists(file string) bool {
	f, err := os.Stat(file)
	return err == nil && f.Mode().IsRegular()
}

func DirExists(dir string) bool {
	d, err := os.Stat(dir)
	return err == nil && d.Mode().IsDir()
}

type Json struct {
	Data map[string]interface{}
}

func (j Json) GetString(key string) string {
	if d, ok := j.Data[key]; ok {
		if d, ok := d.(string); ok {
			return d
		}
	}
	return ""
}

func (j Json) GetStringArray(key string) []string {
	if d, ok := j.GetArray(key); ok {
		array := make([]string, 0)
		for _, value := range d {
			if d, ok := value.(string); ok {
				array = append(array, d)
			} else {
				return nil
			}
		}
		return array
	}
	return nil
}

func (j Json) GetFloatArray(key string) []float64 {
	if d, ok := j.GetArray(key); ok {
		array := make([]float64, 0)
		for _, value := range d {
			if d, ok := value.(float64); ok {
				array = append(array, d)
			} else {
				return nil
			}
		}
		return array
	}
	return nil
}

func (j Json) GetArray(key string) ([]interface{}, bool) {
	if d, ok := j.Data[key]; ok {
		if d, ok := d.([]interface{}); ok {
			return d, true
		}
	}
	return nil, false
}

func (j Json) GetFloat(key string) float64 {
	if d, ok := j.Data[key]; ok {
		if d, ok := d.(float64); ok {
			return d
		}
	}
	return 0
}

func Panic(err error) {
	if err != nil {
		panic(err)
	}
}

func GetAverage(array []float64) float64 {
	var sum float64 = 0

	for _, i := range array {
		sum += i
	}
	sum /= float64(len(array))

	return sum
}

func Encode(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

func Decode(text string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(text)
}

func ToURLBase64(buf []byte) string {
	return base64.URLEncoding.EncodeToString(buf)
}

func FromURLBase64(text string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(text)
}

func InsertToSlice(item string, slice []string, index int) []string {
	var buf []string = make([]string, len(slice)+1)
	copy(buf, slice[:index])
	buf[index] = item
	copy(buf[index+1:], slice[index:])

	return buf
}

func RemoveFromSlice(slice []string, index int) []string {
	var buf []string = make([]string, len(slice)-1)
	copy(buf, slice[:index])
	copy(buf[index:], slice[index+1:])

	return buf
}

type GenericError string

func (message GenericError) Error() string {
	return string(message)
}

func SliceContains(item string, slice []string) bool {
	for _, s := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func FindinSortedList(sortedlist []interface{},
	equals, biggerthan func(i, j interface{}) bool,
	item interface{},
	reversed bool) (int, error) {
	return _findinSortedList(sortedlist, equals,
		biggerthan, item, 0, len(sortedlist)-1,
		reversed)
}

func _findinSortedList(sortedlist []interface{},
	equals, biggerthan func(i, j interface{}) bool,
	item interface{},
	min, max int, reversed bool) (int, error) {

	if len(sortedlist) == 0 {
		return 0, GenericError(fmt.Sprintf("Couldn't find %s", item))
	}

	// Make sure if id actually exists
	// otherwise it will end in an endless loop
	if min >= max {
		if equals(item, sortedlist[min]) {
			return min, nil
		}
		return min, GenericError(fmt.Sprintf("Couldn't find %s", item))
	}

	if max-min == 1 {
		if equals(item, sortedlist[min]) {
			return min, nil
		}
		if equals(item, sortedlist[max]) {
			return max, nil
		}
		if reversed && biggerthan(item, sortedlist[min]) {
			return min, GenericError(fmt.Sprintf("Couldn't find %s", item))
		} else if !reversed && biggerthan(item, sortedlist[max]) {
			return max, GenericError(fmt.Sprintf("Couldn't find %s", item))
		} else {
			if reversed {
				return max, GenericError(fmt.Sprintf("Couldn't find %s", item))
			} else {
				return min, GenericError(fmt.Sprintf("Couldn't find %s", item))
			}
		}
	}

	index := (max-min)/2 + min
	middleItem := sortedlist[index]

	searchleft := true
	if biggerthan(item, middleItem) {
		searchleft = false
	} else if equals(item, middleItem) {
		return index, nil
	}

	if searchleft != reversed {
		return _findinSortedList(sortedlist, equals, biggerthan,
			item, min, index-1, reversed)

	}
	return _findinSortedList(sortedlist, equals, biggerthan,
		item, index+1, max, reversed)
}
