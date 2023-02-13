// RGL api implementation
// See README for usage and gotchas.
package rgl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"strings"
	"time"
)

const RGL_ENDPOINT = "https://api.rgl.gg/v0/"
const PLAYER_ENDPOINT = RGL_ENDPOINT + "profile/"
const TEAM_ENDPOINT = RGL_ENDPOINT + "teams/"
const SEASON_ENDPOINT = RGL_ENDPOINT + "seasons/"
const SEARCH_ALIAS_ENDPOINT = RGL_ENDPOINT + "search/players"
const BULK_PLAYER_ENDPOINT = PLAYER_ENDPOINT + "getmany"
const MATCH_ENDPOINT = RGL_ENDPOINT + "matches/"
const SEARCH_TEAM_ENDPOINT = RGL_ENDPOINT + "search/teams"

// Used in Player.CurrentTeams to represent the teams a player is on
type CurrTeam struct {
	Id       int    `json:"id"`
	Tag      string `json:"tag"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	SeasonId int    `json:"seasonId"`
	DivId    int    `json:"divisionId"`
	DivName  string `json:"divisionName"`
}

// Used in Player to represent status information
type PlayerStatus struct {
	IsVerified    bool `json:"isVerified"`
	IsBanned      bool `json:"isBanned"`
	IsOnProbation bool `json:"isOnProbation"`
}

// Used in Player to represent ban information
type Ban struct {
	Ends   string `json:"endsAt"`
	Reason string `json:"reason"`
}

// Used when receiving paginated bans
type BulkBan struct {
	SteamId string `json:"steamId"`
	Alias   string `json:"alias"`
	Expires string `json:"expiresAt"`
	Created string `json:"createdAt"`
	Reason  string `json:"reason"`
}

// Used in Player to represent the teams a player is on in each format
type CurrentTeams struct {
	Sixes      *CurrTeam `json:"sixes"`
	Highlander *CurrTeam `json:"highlander"`
	Prolander  *CurrTeam `json:"prolander"`
}

// Toplevel endpoint for a player found by id
type Player struct {
	SteamId      string       `json:"steamId"`
	Avatar       string       `json:"avatar"`
	Name         string       `json:"name"`
	Updated      string       `json:"updatedAt"`
	Status       PlayerStatus `json:"status"`
	Ban          *Ban         `json:"banInformation"`
	CurrentTeams CurrentTeams `json:"currentTeams"`
}

// To check if this is empty, use len(SearchResults.Results) == 0 instead of equality checking == SearchResults{}
// (If you want to equality check the results of rgl.SearchPlayers, you'd have to use SearchResults{Results: make([]string, 0)}
type SearchResults struct {
	Results       []string `json:"results"` //A slice of Steam64 IDs (Steam64s for Players, RGL Team IDs for Teams)
	Count         int      `json:"int"`
	TotalHitCount int      `json:"totalHitCount"`
}

// Used in Team to represent at-a-glance information about the roster
type TeamPlayer struct {
	Name     string `json:"name"`
	SteamId  string `json:"steamId"`
	IsLeader bool   `json:"isLeader"`
	Joined   string `json:"joinedAt"`
}

// Toplevel endpoint for a team found by id
type Team struct {
	Id          int          `json:"teamId"`
	LinkedTeams []int        `json:"linkedTeams"`
	SeasonId    int          `json:"seasonId"`
	DivId       int          `json:"divisionId"`
	DivName     string       `json:"divisionName"`
	TeamLeader  string       `json:"teamLeader"`
	Created     string       `json:"createdAt"`
	Updated     string       `json:"updatedAt"`
	Tag         string       `json:"tag"`
	Name        string       `json:"name"`
	FinalRank   *int         `json:"finalRank"`
	Players     []TeamPlayer `json:"players"`
}

// Used when finding the past teams of a specific player ID
type PlayerTeamHistory struct {
	FormatId     int    `json:"formatId"`
	FormatName   string `json:"formatName"`
	RegionId     int    `json:"regionId"`
	RegionName   string `json:"regionName"`
	SeasonId     int    `json:"seasonId"`
	SeasonName   string `json:"seasonName"`
	Started      string `json:"startedAt"`
	DivisionId   int    `json:"divisionId"`
	DivisionName string `json:"divisionName"`
	Left         string `json:"leftAt"` //Can be empty to indicate a null value (player hasn't left team)
	TeamName     string `json:"teamName"`
	TeamTag      string `json:"teamTag"`
	TeamId       int    `json:"teamId"`
	Stats        struct {
		Wins         int `json:"wins"`
		WinsWithout  int `json:"winsWithout"`
		Loses        int `json:"loses"`
		LosesWithout int `json:"losesWithout"`
		GamesPlayed  int `json:"gamesPlayed"`
		GamesWithout int `json:"gamesWithout"`
	} `json:"stats"`
}

// Embedded in Match to represent the teams playing
type MatchTeam struct {
	Id       int    `json:"teamId"`
	TeamName string `json:"teamName"`
	TeamTag  string `json:"teamTag"`
	IsHome   bool   `json:"isHome"`
	Points   string `json:"points"`
}

// Embedded in Match to represent maps played (multiple maps per match for playoffs)
type MatchMap struct {
	MapName   string `json:"mapName"`
	HomeScore int    `json:"homeScore"`
	AwayScore int    `json:"awayScore"`
}

// Toplevel endpoint for a season found by id
type Season struct {
	Name    string   `json:"name"`
	Format  *string  `json:"formatName"`
	Region  *string  `json:"regionName"`
	Maps    []string `json:"maps"`
	Teams   []int    `json:"participatingTeams"`
	Matches []int    `json:"matchesPlayedDuringSeason"`
}

// Toplevel endpoint for a match found by id
type Match struct {
	Id         int         `json:"matchId"`
	SeasonName string      `json:"seasonName"`
	DivName    string      `json:"divName"`
	SeasonId   int         `json:"seasonId"`
	MatchDate  string      `json:"matchDate"`
	MatchName  string      `json:"matchName"`
	Teams      []MatchTeam `json:"teams"`
	Maps       []MatchMap  `json:"maps"`
}

// An error that comes from POSTing (at least to /search/players. Message has varying fields, but the important ones are constant.
type PostError struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"message"`
}

// The RGL type contains all endpoints as methods. Create one with rgl.DefaultRateLimit()
// or use RGL{rl: *rate.Limiter}
// or use RGL{} if you don't want to use the ratelimiter object at all (you will have to implement your own)
type RGL struct {
	rl *rate.Limiter
}

// Create an RGL instance with a default rate limiter based on present ratelimits (2 calls per 1 second)>
func DefaultRateLimit() RGL {
	r := RGL{}
	r.rl = rate.NewLimiter(rate.Every(time.Second), 2) //2 requests every second
	return r
}

// Wrapper around time.Parse("2006-01-02T15:04:05.999Z", str)
func ToGoTime(str string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.999Z", str)
	return t
}

func (rgl *RGL) get(url string) (io.ReadCloser, error) {
	if rgl.rl != nil { //If using the pkgs ratelimiter (user should implement their own if they don't want to use the default)
		ctx := context.Background()
		err := rgl.rl.Wait(ctx)

		if err != nil {
			return nil, fmt.Errorf("Error waiting on ratelimiter %v\n", err)
		}
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error getting endpoint %s: %v\n", url, err)
	}
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("Hit ratelimit")
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Not Found")
	}
	return resp.Body, nil //resp.Body is not closed here. Defer it after calling get
}

func (rgl *RGL) post(url string, body interface{}) (*http.Response, error) {
	if rgl.rl != nil {
		ctx := context.Background()
		err := rgl.rl.Wait(ctx)

		if err != nil {
			return nil, fmt.Errorf("Error waiting on ratelimiter %v\n", err)
		}
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling request body: %v\n", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	//It seems RGL's error API may vary based on the actual endpoint, so I'm going to leave error handling to further down the func chain.
	return resp, nil
}

// Get player by steam id
func (rgl *RGL) GetPlayer(steam64 string) (Player, error) {
	var p Player
	if !strings.HasPrefix(steam64, "765611") {
		return p, fmt.Errorf("Steam64 must begin with 765611")
	}
	url := PLAYER_ENDPOINT + steam64
	body, err := rgl.get(url)
	if err != nil {
		if err.Error() == "Not Found" {
			return p, nil
		}
		return p, fmt.Errorf("Error getting player: %v", err)
	}
	defer body.Close()
	err = json.NewDecoder(body).Decode(&p)
	if err != nil {
		return p, fmt.Errorf("Error decoding json response: %v", err)
	}
	return p, nil
}

// Get team by RGL Id
func (rgl *RGL) GetTeam(id int) (Team, error) {
	var t Team
	url := TEAM_ENDPOINT + fmt.Sprint(id)
	body, err := rgl.get(url)
	if err != nil {
		if err.Error() == "Not Found" {
			return t, nil
		}
		return t, fmt.Errorf("Error getting team: %v", err)
	}
	defer body.Close()
	err = json.NewDecoder(body).Decode(&t)
	if err != nil {
		return t, fmt.Errorf("Error decoding json response: %v", err)
	}
	return t, nil
}

// Get season by RGL Id
func (rgl *RGL) GetSeason(id int) (Season, error) {
	var s Season
	url := SEASON_ENDPOINT + fmt.Sprint(id)
	body, err := rgl.get(url)
	if err != nil {
		if err.Error() == "Not Found" {
			return s, nil
		}
		return s, fmt.Errorf("Error getting season: %v", err)
	}
	defer body.Close()
	err = json.NewDecoder(body).Decode(&s)
	if err != nil {
		return s, fmt.Errorf("Error decoding json response: %v", err)
	}
	return s, nil
}

// Search for players whose aliases contain the string. Take the first `take` results, skipping the first `skip`.
func (rgl *RGL) SearchPlayers(alias string, take int, skip int) (SearchResults, error) {
	var results SearchResults
	if len(alias) < 2 {
		return results, fmt.Errorf("Length of alias must be at least 2")
	}
	url := fmt.Sprintf("%s?take=%d&skip=%d", SEARCH_ALIAS_ENDPOINT, take, skip)
	resp, err := rgl.post(url, struct {
		NameContains string `json:"nameContains"`
	}{alias})
	if err != nil {
		return results, fmt.Errorf("Error POSTing for player aliases: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var pe PostError
		err = json.NewDecoder(resp.Body).Decode(&pe)
		if err != nil {
			if err.Error() == "Not Found" {
				return results, nil
			}
			return results, err
		}
		erCode := pe.Message[0].Code
		if erCode == "invalid_type" { //Neither of these should occur if the library operates properly
			return results, fmt.Errorf("Library error (invalid_type)") //nameContains encoded incorrectly
		} else if erCode == "too_small" {
			return results, fmt.Errorf("Alias too short") //len(alias) < 2 check should make this redundant
		}
		return results, fmt.Errorf("PostError body: %v", pe)
	}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return results, fmt.Errorf("Error decoding json response: %v", err)
	}
	return results, nil
}

// Search multiple IDs for RGL players
func (rgl *RGL) BulkPlayers(ids []string) ([]Player, error) {
	players := make([]Player, 0)
	url := BULK_PLAYER_ENDPOINT
	resp, err := rgl.post(url, ids)
	if err != nil {
		return players, fmt.Errorf("Error POSTing for bulk players: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(&players)
		return players, err //players will be the empty slice declared at the top if err is not nil
	} else if resp.StatusCode == 404 {
		return players, nil
	} else { //statuscode is technically 400 but returns a json PostError.StatusCode = 404
		var pe PostError
		err = json.NewDecoder(resp.Body).Decode(&pe)
		if err != nil {
			return players, err
		}
		erCode := pe.Message[0].Code
		if erCode == "invalid_string" {
			return players, fmt.Errorf("One or more steamids was invalid")
		}
		return nil, fmt.Errorf("PostError body: %v", pe)
	}
}

// Get match by RGL Id
func (rgl *RGL) GetMatch(id int) (Match, error) {
	var m Match
	url := MATCH_ENDPOINT + fmt.Sprint(id)
	body, err := rgl.get(url)
	if err != nil {
		if err.Error() == "Not Found" {
			return m, nil
		}
		return m, fmt.Errorf("Error getting match: %v", err)
	}
	defer body.Close()
	err = json.NewDecoder(body).Decode(&m)
	if err != nil {
		return m, fmt.Errorf("Error decoding json response: %v", err)
	}
	return m, nil
}

// Bulk search for teams whose names or tags contain the partial string.
func (rgl *RGL) SearchTeams(partial string, take int, skip int) (SearchResults, error) {
	var results SearchResults
	if len(partial) < 2 {
		return results, fmt.Errorf("Length of partial string must be at least 2")
	}
	url := fmt.Sprintf("%s?take=%d&skip=%d", SEARCH_TEAM_ENDPOINT, take, skip)
	resp, err := rgl.post(url, struct {
		NameContains string `json:"nameContains"`
	}{partial})
	if err != nil {
		return results, fmt.Errorf("Error POSTing for team bulk: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var pe PostError
		err = json.NewDecoder(resp.Body).Decode(&pe)
		if err != nil {
			if err.Error() == "Not Found" {
				return results, nil
			}
			return results, err
		}
		erCode := pe.Message[0].Code
		if erCode == "invalid_type" { //Neither of these should occur if the library operates properly
			return results, fmt.Errorf("Library error (invalid_type)") //nameContains encoded incorrectly
		} else if erCode == "too_small" {
			return results, fmt.Errorf("Alias too short") //len(alias) < 2 check should make this redundant
		}
		return results, fmt.Errorf("PostError body: %v", pe)
	}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return results, fmt.Errorf("Error decoding json response: %v", err)
	}
	return results, nil
}

// Get a player's teams (past and present). Current teams have the Left field as ""
func (rgl *RGL) GetPlayerTeamHistory(id string) ([]PlayerTeamHistory, error) {
	teams := make([]PlayerTeamHistory, 0)
	url := PLAYER_ENDPOINT + id + "/teams"
	body, err := rgl.get(url)
	if err != nil {
		if err.Error() == "Not Found" {
			return teams, nil
		}
		return teams, fmt.Errorf("Error getting team history")
	}
	defer body.Close()
	err = json.NewDecoder(body).Decode(&teams)
	if err != nil {
		return teams, err
	}
	return teams, nil
}

// A paginated look at RGL bans. (Newest first). This is a historic record and includes expired bans, as far as I can tell.
func (rgl *RGL) GetBans(take int, skip int) ([]BulkBan, error) {
	bans := make([]BulkBan, 0)
	url := fmt.Sprintf("https://api.rgl.gg/v0/bans/paged?take=%d&skip=%d", take, skip)
	body, err := rgl.get(url)
	if err != nil {
		return bans, fmt.Errorf("Error getting paginated bans")
	}
	defer body.Close()
	err = json.NewDecoder(body).Decode(&bans)
	if err != nil {
		return bans, err
	}
	return bans, nil
}
