package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type datat struct {
	dnsAddr   string
	domain    string
	domainID  string
	id        string
	loginAddr string
	password  string
	username  string
}

func getCurIP() (string, error) {
	res, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}

	newIP, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(newIP), nil
}

func updateIP(data datat, ip string) error {

	// Our client with its cookies
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	// Login form data
	postData := url.Values{}
	postData.Add("UserForm[email]", data.username)
	postData.Add("UserForm[password]:", data.password)
	postData.Add("UserForm[login_duration_flag]", "1")

	// Perform a login in order to get our credentials cookie
	login, err := client.PostForm(data.loginAddr, postData)
	if err != nil {
		return err
	}

	// There should be a better way to do that
	mockJSON := `{"id":"123","domain_id":"456","name":"home","type":"A","content":"1.1.1.1","ttl":1800,"prio":null,"change_date":null,"disabled":"0","ordername":null,"auth":"1","description":"Host","selected_ttl":{"name":"30 minutes","value":1800}}`

	var jsonMap map[string]interface{}

	json.Unmarshal([]byte(mockJSON), &jsonMap)

	// Our relevant data to update our IP
	jsonMap["id"] = data.id
	jsonMap["domain_id"] = data.domainID
	jsonMap["content"] = ip

	formattedJSON, err := json.Marshal(jsonMap)
	if err != nil {
		return err
	}

	// Actual request to update the IP
	req, _ := http.NewRequest("PUT", data.dnsAddr, strings.NewReader(string(formattedJSON)))
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	defer login.Body.Close()

	return nil
}

func main() {
	envPtr := flag.String("env", "", "Custom env file location")

	flag.Parse()

	var envVars map[string]string
	var err error
	if *envPtr == "" {
		envVars, err = godotenv.Read()
	} else {
		envVars, err = godotenv.Read(*envPtr)
	}

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	data := datat{
		dnsAddr:   envVars["DNSADDR"],
		domain:    envVars["DOMAIN"],
		domainID:  envVars["DOMAINID"],
		id:        envVars["ID"],
		loginAddr: envVars["LOGINADDR"],
		password:  envVars["PASSWORD"],
		username:  envVars["USERNAME"],
	}

	ip := ""

	for {

		newIP, err := getCurIP()
		if err != nil {
			log.Println("[WARN]", err)
		} else {
			if ip != newIP {

				if err := updateIP(data, newIP); err != nil {
					log.Println("[WARN]", err)
				} else {
					log.Println("[INFO] IP updated to", newIP)
					ip = newIP
				}

			}
		}

		// Sleeps for 5 minutes before checking for changes again
		time.Sleep(5 * time.Minute)
	}
}
