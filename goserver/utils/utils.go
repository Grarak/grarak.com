package utils

import (
	"encoding/base64"
	"os"
	"sort"
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

type sorter struct {
	array              []string
	minmaxDeterminator func(i, j int) bool
}

func (sorter sorter) Len() int {
	return len(sorter.array)
}

func (sorter sorter) Swap(i, j int) {
	sorter.array[i], sorter.array[j] = sorter.array[j], sorter.array[i]
}

func (sorter sorter) Less(i, j int) bool {
	return sorter.minmaxDeterminator(i, j)
}

func SimpleSort(array []string, minmaxDeterminator func(i, j int) bool) {
	sort.Sort(sorter{array, minmaxDeterminator})
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
