package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	rtctokenbuilder "github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/rtctokenbuilder2"
)

type rtc_int_token_struct struct {
	Uid_rtc_int  uint32 `json:"uid"`
	Channel_name string `json:"ChannelName"`
	Role         uint32 `json:"role"`
}

var rtc_token string
var int_uid uint32
var channel_name string

var role_num uint32
var role rtctokenbuilder.Role

// Use RtcTokenBuilder to generate an <Vg k="VSDK" /> token.
func generateRtcToken(int_uid uint32, channelName string, role rtctokenbuilder.Role) {

	appID := "92d7c7017d4d494f9b4a8d2649b3bd1a"
	appCertificate := "b2bcbf8eed294be7a47c2a4cddbc960c"
	// Number of seconds after which the AccessToken2 expires.
	// When the AccessToken2 expires but the privilege does not expire, the user remains in the channel and can continue to publish streams. No callback is triggered from the SDK.
	// However, once disconnected from the channel, the user cannot rejoin the channel with that token. Ensure the AccessToken2 does not expire before the privileges.
	tokenExpireTimeInSeconds := uint32(40)
	// Number of seconds after which the privilege expires.
	// The token-privilege-will-expire callback occurs 30 seconds before the privilege expires.
	// The token-privilege-did-expire callback occurs when the privilege expires.
	// For demonstration purposes the expire time is set to 40 seconds. This shows you the automatic token renew actions of the client.
	privilegeExpireTimeInSeconds := uint32(40)

	result, err := rtctokenbuilder.BuildTokenWithUid(appID, appCertificate, channelName, int_uid, role, tokenExpireTimeInSeconds, privilegeExpireTimeInSeconds)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Token with uid: %s\n", result)
		fmt.Printf("uid is %d\n", int_uid)
		fmt.Printf("ChannelName is %s\n", channelName)
		fmt.Printf("Role is %d\n", role)
	}
	rtc_token = result
}

func rtcTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" && r.Method != "OPTIONS" {
		http.Error(w, "Unsupported method. Please check.", http.StatusNotFound)
		return
	}

	var t_int rtc_int_token_struct
	var unmarshalErr *json.UnmarshalTypeError
	int_decoder := json.NewDecoder(r.Body)
	int_err := int_decoder.Decode(&t_int)
	if int_err == nil {

		int_uid = t_int.Uid_rtc_int
		channel_name = t_int.Channel_name
		role_num = t_int.Role
		switch role_num {
		case 1:
			role = rtctokenbuilder.RolePublisher
		case 2:
			role = rtctokenbuilder.RoleSubscriber
		}
	}
	if int_err != nil {

		if errors.As(int_err, &unmarshalErr) {
			errorResponse(w, "Bad request. Wrong type provided for field "+unmarshalErr.Value+unmarshalErr.Field+unmarshalErr.Struct, http.StatusBadRequest)
		} else {
			errorResponse(w, "Bad request.", http.StatusBadRequest)
		}
		return
	}

	generateRtcToken(int_uid, channel_name, role)
	errorResponse(w, rtc_token, http.StatusOK)
	log.Println(w, r)
}

func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["token"] = message
	resp["code"] = strconv.Itoa(httpStatusCode)
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)

}

func main() {
	// <Vg k="VSDK" /> token from <Vg k="VSDK" /> int uid
	http.HandleFunc("/fetch_rtc_token", rtcTokenHandler)
	fmt.Printf("Starting server at port 8080\n")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("server deployed")
		log.Fatal(err)
	}
}
