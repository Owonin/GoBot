package commands

import (
	"bytes"
	"discord-bot/config"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	token_Url string = "http://osu.ppy.sh/oauth/token"
	api_Url   string = "http://osu.ppy.sh/api/v2"
	token     string = ""
)

type Result struct {
	Result string `json:"access_token"`
}

type OsuUser struct {
	Discord       string      `json:"discord"`
	Join_date     time.Time   `json:"Timestamp "`
	Playmode      string      `json:"playmode"` //todo make gamemodestruct
	Playstyle     []string    `json:"playstyle"`
	Post_count    int         `json:"post_count"`
	Profile_order ProfilePage `json:"profile_order"`
	Twitter       string      `json:"twitter"`
	Website       string      `json:"website"`
}

type OsuUserDetail struct {
	*OsuUser
	Avatar_url    string     `json:"avatar_url"`
	Country_code  string     `json:"country_code"`
	Id            int        `json:"id"`
	Is_active     bool       `json:"is_active"`
	Is_bot        bool       `json:"is_bot"`
	Is_deleted    bool       `json:"is_deleted"`
	Is_online     bool       `json:"is_online"`
	Last_visit    time.Time  `json:"last_visit"`
	Username      string     `json:"username"`
	GradeCounts   string     `json:"grade_counts"`
	UserStatistic Statistics `json:"statistics"`
}

type Statistics struct {
	GlobalRank  uint    `json:"global_rank"`
	Pp          float32 `json:"pp"`
	HitAccuracy float32 `json:"hit_accuracy"`
	Playtime    uint    `json:"play_time"`
	MaxCombo    uint    `json:"maximum_combo"`
	// GradeCnt       GradeCounts `json:"grade_counts"` TODO add Gcnts
	CountryRank uint `json:"country_rank"`
}

type ProfilePage struct {
	Me              []string `json:"me"`
	Recent_activity []string `json:"recent_activity"`
	Beatmaps        []string `json:"beatmaps"`
	Historical      []string `json:"historical"`
	Kudosu          []string `json:"kudosu"`
	Top_ranks       []string `json:"top_ranks"`
	Medals          []string `json:"medals"`
}

func GetOsuToken() (string, error) {

	values := map[string]string{"client_id": config.OsuClientId, "client_secret": config.OsuToken, "grant_type": "client_credentials", "scope": "public"}

	jsonValue, _ := json.Marshal(values)

	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, token_Url, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Fatal("Client creatign error")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var result Result
	json.Unmarshal(body, &result)
	return result.Result, nil

}

func GetOsuUser(name string) *OsuUserDetail {

	client := http.Client{}
	req, err := http.NewRequest("GET", api_Url+"/users/"+name+"/osu", nil)
	if err != nil {
		log.Fatal("Client creatign error")
	}
	if token == "" {
		token, err = GetOsuToken()
		if err != nil {
			log.Panic(err)
		}
	}

	req.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Accept":        []string{"application/json"},
		"Authorization": []string{"Bearer " + token},
	}

	q := req.URL.Query()
	q.Add("limit", "5")

	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		log.Fatal("Osu request error")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, body, "", "\t")
	if error != nil {
		log.Println("JSON parse error: ", error)
	}

	log.Println("CSP Violation:", prettyJSON.String())

	var result OsuUserDetail
	json.Unmarshal(body, &result)
	return &result
}
