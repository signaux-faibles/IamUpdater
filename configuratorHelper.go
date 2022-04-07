package main

type Configurator interface {
	GetConfig() map[string]interface{}
}

func updateBoolParam(cc Configurator, key string, setter func(*bool)) {
	if val, ok := fetchParam(cc, key); ok {
		t := val.(bool)
		setter(&t)
	}
}

func updateIntParam(cc Configurator, key string, setter func(param *int)) {
	if val, ok := fetchParam(cc, key); ok {
		r := int(val.(int64))
		setter(&r)
	}
}

func updateMapOfStringsParam(cc Configurator, key string, setter func(*map[string]string)) {
	if val, ok := fetchParam(cc, key); ok {
		r := map[string]string{}
		t := val.(map[string]interface{})
		for k, v := range t {
			r[k] = v.(string)
		}
		setter(&r)
	}
}

func updateStringParam(cc Configurator, key string, setter func(*string)) {
	if val, ok := fetchParam(cc, key); ok {
		r := val.(string)
		setter(&r)
	}
}

func updateStringArrayParam(cc Configurator, key string, setter func(*[]string)) {
	if val, ok := fetchParam(cc, key); ok {
		vals := val.([]interface{})
		var t []string
		for _, uri := range vals {
			t = append(t, uri.(string))
		}
		setter(&t)
	}
}

func fetchParam(cc Configurator, key string) (interface{}, bool) {
	config := cc.GetConfig()
	val, ok := config[key]
	if !ok {
		return nil, false
	}
	delete(config, key)
	return val, true
}
