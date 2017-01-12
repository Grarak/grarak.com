package utils

import (
	"encoding/base64"
	"os"
)

func StringEmpty(text string) bool {
	return len(text) == 0
}

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
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
