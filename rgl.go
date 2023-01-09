// RGL api implementation
// Coverage:
// matches: x
//
// profile:
//	by id: o
//	by name: x
//	past teams: x
//
// teams:
//	by id: o
//	by name: x
//
// seasons: o
//
// bans: x
package rgl

import (
	"net/http"
	"golang.org/x/time/rate"
	"context"
	"fmt"
	"time"
	"encoding/json"
	"io"
)

const RGL_ENDPOINT = "https://api.rgl.gg/v0/"
const PLAYER_ENDPOINT = RGL_ENDPOINT + "profile/"
const TEAM_ENDPOINT = RGL_ENDPOINT + "teams/"
const SEASON_ENDPOINT = RGL_ENDPOINT + "seasons/"

// Used in Player.CurrentTeams to represent the teams a player is on
type CurrTeam struct {
	Id int `json:"id"`
	Tag string `json:"tag"`
	Name string `json:"name"`
	Status string `json:"status"`
	SeasonId int `json:"seasonId"`
	DivId int `json:"divisionId"`
	DivName int `json:"divisionName"`
}

// Used in Player to represent status information
type PlayerStatus struct {
	IsVerified bool `json:"isVerified"`
	IsBanned bool `json:"isBanned"`
	IsOnProbation bool `json:"isOnProbation"`
}

// Used in Player to represent ban information
type Ban struct {
	Ends string `json:"endsAt"`
	Reason string `json:"reason"`
}

// Used in Player to represent the teams a player is on in each format
type CurrentTeams struct {
	Sixes *CurrTeam `json:"sixes"`
	Highlander *CurrTeam `json:"highlander"`
	Prolander *CurrTeam `json:"prolander"`
}

// Toplevel endpoint for a player found by id
type Player struct {
	SteamId string `json:"steamId"`
	Avatar string `json:"avatar"`
	Name string `json:"name"`
	Updated string `json:"updatedAt"`
	Status PlayerStatus `json:"status"`
	Ban *Ban `json:"banInformation"`
	CurrentTeams CurrentTeams `json:"currentTeams"`
}

// Used in Team to represent at-a-glance information about the roster
type TeamPlayer struct {
	Name string `json:"name"`
	SteamId string `json:"steamId"`
	IsLeader bool `json:"isLeader"`
	Joined string `json:"joinedAt"`
}

// Toplevel endpoint for a team found by id
type Team struct {
	Id int `json:"teamId"`
	LinkedTeams []int `json:"linkedTeams"`
	SeasonId int `json:"seasonId"`
	DivId int `json:"divisionId"`
	DivName string `json:"divisionName"`
	TeamLeader string `json:"teamLeader"`
	Created string `json:"createdAt"`
	Updated string `json:"updatedAt"`
	Tag string `json:"tag"`
	Name string `json:"name"`
	FinalRank *int `json:"finalRank"`
	Players []TeamPlayer `json:"players"`
}

// Toplevel endpoint for a season found by id
type Season struct {
	Name string `json:"name"`
	Format *string `json:"formatName"`
	Region *string `json:"regionName"`
	Maps []string `json:"maps"`
	Teams []int `json:"participatingTeams"`
	Matches []int `json:"matchesPlayedDuringSeason"`
}

// The RGL type contains all endpoints as methods. Create one with rgl.DefaultRateLimit()
// or use RGL{rl: *rate.Limiter} if you know what you're doing
type RGL struct {
	rl *rate.Limiter
}

// Create an RGL instance with a default rate limiter based on present ratelimits (2 calls per 1 second)>
func DefaultRateLimit() RGL {
	r := RGL{}
	r.rl = rate.NewLimiter(rate.Every(time.Second), 2) //2 requests every second
	return r
}

func (rgl *RGL) get(url string) (io.ReadCloser, error) {
	ctx := context.Background()
	err := rgl.rl.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error waiting on ratelimiter %v\n", err)
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error getting endpoint %s: %v\n", url, err)
	}
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("Hit ratelimit")
	}
	return resp.Body, nil //resp.Body is not closed here. Defer it after calling get
}

func (rgl *RGL) GetPlayer(steam64 string) (*Player, error) {
	url := PLAYER_ENDPOINT + steam64
	body, err := rgl.get(url)
	if err != nil {
		return nil, fmt.Errorf("Error getting player: %v", err)
	}
	defer body.Close()
	var p Player //We are not declaring as a pointer because for some reason you can't just do Decode(p).
	err = json.NewDecoder(body).Decode(&p)
	if err != nil {
		return nil, fmt.Errorf("Error decoding json response: %v", err)
	}
	return &p, nil //We are returning the pointer so we have the option of nil-ing.
}

func (rgl *RGL) GetTeam(id int) (*Team, error) {
	url := TEAM_ENDPOINT + fmt.Sprint(id)
	body, err := rgl.get(url)
	if err != nil {
		return nil, fmt.Errorf("Error getting team: %v", err)
	}
	defer body.Close()
	var t Team
	err = json.NewDecoder(body).Decode(&t)
	if err != nil {
		return nil, fmt.Errorf("Error decoding json response: %v", err)
	}
	return &t, nil
}

func (rgl *RGL) GetSeason(id int) (*Season, error) {
	url := SEASON_ENDPOINT + fmt.Sprint(id)
	body, err := rgl.get(url)
	if err != nil {
		return nil, fmt.Errorf("Error getting season: %v", err)
	}
	defer body.Close()
	var s Season
	err = json.NewDecoder(body).Decode(&s)
	if err != nil {
		return nil, fmt.Errorf("Error decoding json response: %v", err)
	}
	return &s, nil
}