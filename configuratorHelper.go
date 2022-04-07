package main

type Configurator interface {
	GetConfig() map[string]interface{}
}

func updateBoolParam(cc Configurator, key string, setter func(*bool)) {
	config := cc.GetConfig()
	val, ok := config[key]
	if !ok {
		return
	}
	delete(config, key)
	t := val.(bool)
	setter(&t)
}

func updateIntParam(cc Configurator, key string, setter func(param *int)) {
	config := cc.GetConfig()
	val, ok := config[key]
	if !ok {
		return
	}
	delete(config, key)
	r := int(val.(int64))
	setter(&r)
}

func updateMapOfStringsParam(cc Configurator, key string, setter func(*map[string]string)) {
	config := cc.GetConfig()
	val, ok := config[key]
	if !ok {
		return
	}
	delete(config, key)
	r := map[string]string{}
	t := val.(map[string]interface{})
	for k, v := range t {
		r[k] = v.(string)
	}
	setter(&r)
}

func updateStringParam(cc Configurator, key string, setter func(*string)) {
	config := cc.GetConfig()
	val, ok := config[key]
	if !ok {
		return
	}
	delete(config, key)
	r := val.(string)
	setter(&r)
}

func updateStringArrayParam(cc Configurator, key string, setter func(*[]string)) {
	config := cc.GetConfig()
	val, ok := config[key]
	if !ok {
		return
	}
	delete(config, key)
	vals := val.([]interface{})
	var t []string
	for _, uri := range vals {
		t = append(t, uri.(string))
	}
	setter(&t)
}
