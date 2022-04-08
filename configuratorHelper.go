package main

import "log"

type Configurator interface {
	GetConfig() map[string]interface{}
}

func updateBoolParam(configurator Configurator, key string, setter func(*bool)) {
	if val, ok := fetchParam(configurator, key); ok {
		t := val.(bool)
		setter(&t)
	}
}

func updateIntParam(configurator Configurator, key string, setter func(param *int)) {
	if val, ok := fetchParam(configurator, key); ok {
		r := int(val.(int64))
		setter(&r)
	}
}

func updateMapOfStringsParam(configurator Configurator, key string, setter func(*map[string]string)) {
	if val, ok := fetchParam(configurator, key); ok {
		r := map[string]string{}
		t := val.(map[string]interface{})
		for k, v := range t {
			r[k] = v.(string)
		}
		setter(&r)
	}
}

func updateStringParam(configurator Configurator, key string, setter func(*string)) {
	if val, ok := fetchParam(configurator, key); ok {
		r := val.(string)
		setter(&r)
	}
}

func updateStringArrayParam(configurator Configurator, key string, setter func(*[]string)) {
	if val, ok := fetchParam(configurator, key); ok {
		vals := val.([]interface{})
		var t []string
		for _, uri := range vals {
			t = append(t, uri.(string))
		}
		setter(&t)
	}
}

func fetchParam(configurator Configurator, key string) (interface{}, bool) {
	config := configurator.GetConfig()
	val, ok := config[key]
	if !ok {
		log.Printf("%s - param '%s' is not found", configurator, key)
		return nil, false
	}
	delete(config, key)
	return val, true
}
