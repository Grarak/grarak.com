package utils

import "os"

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
	if d, ok := j.Data[key]; ok {
		if d, ok := d.([]interface{}); ok {
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
	}
	return nil
}

func (j Json) GetFloatArray(key string) []float64 {
	if d, ok := j.Data[key]; ok {
		if d, ok := d.([]interface{}); ok {
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
	}
	return nil
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
