package hosts

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
)

const CONFIG_PATH = "./config.yml"

var hosts = make(map[string]string)

//var hosts = map[string]string{
//	"localhost":                  "127.0.0.1:2000",
//	"kubernetes.docker.internal": "127.0.0.1:3000",
//}

func GetHost(domain string) (string, error) {
	if host, ok := hosts[domain]; ok {
		return host, nil
	}
	return "", errors.New("host not found")
}

func LoadConfig() {
	f, err := os.Open(CONFIG_PATH)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&hosts)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func SaveConfig(newConfig map[string]string) {
	hosts = newConfig
	f, err := os.Open(CONFIG_PATH)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	err = encoder.Encode(hosts)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func AcceptNewConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	SaveConfig(data)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Successfully Updated Config"})
}
