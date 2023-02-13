package rgl

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var r = DefaultRateLimit()

func TestGetPlayer(t *testing.T) {
	//First test: All-around struct comparison
	expected := Player{
		SteamId: "76561198098770013",
		Avatar:  "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/81/8148c37b434814fb7a4bc175a608c1353b4d0a11_full.jpg",
		Name:    "Captain Zidgel",
		Updated: "2023-02-12T21:48:27.196Z",
		Status: PlayerStatus{
			IsVerified:    false,
			IsBanned:      false,
			IsOnProbation: false,
		},
		Ban: nil,
		CurrentTeams: CurrentTeams{
			Sixes:      nil,
			Highlander: nil,
			Prolander:  nil,
		},
	}

	got, err := r.GetPlayer("76561198098770013")
	require.NoError(t, err, "Shouldn't get error querying for player")
	require.Equal(t, expected, got, "Should have equal player structs")

	//Second test: Just verify some other fields not on the other player ID (I didn't want to rewrite the struct literal)
	expected_ban := &Ban{
		Ends:   "9999-08-24T06:20:00.000Z",
		Reason: "Old account, new account: <a href=\"https://rgl.gg/Public/PlayerProfile.aspx?p=76561198113990147\">https://rgl.gg/Public/PlayerProfile.aspx?p=76561198113990147</a>\r\n</br></br>\r\n(10/9/2021) - Failure to Submit Demos: 1st Offense",
	}

	got, err = r.GetPlayer("76561198011940487")
	require.NoError(t, err, "Shouldn't get error querying for player")
	require.Equal(t, expected_ban, got.Ban, "Should have equal bans")

	//before, _, _ := strings.Cut(got.Ban.Ends, "T")
	tim := ToGoTime(got.Ban.Ends)
	expect_time := time.Date(9999, 8, 24, 6, 20, 0, 0, time.UTC)
	require.Equal(t, expect_time, tim, "Should have equal ban times")

	_, err = r.GetPlayer("unvalidated12345")
	require.Error(t, err)

	got, err = r.GetPlayer("7656111111111111111111111111")
	require.NoError(t, err)
	require.Equal(t, Player{}, got, "Should get empty object for 404")
}

func TestGetTeam(t *testing.T) {
	str := `{
  "teamId": 5979,
  "linkedTeams": [],
  "seasonId": 67,
  "divisionId": 78,
  "divisionName": "Intermediate",
  "teamLeader": "76561198116072296",
  "createdAt": "2020-01-06T00:37:16.236Z",
  "updatedAt": "2021-05-26T01:36:52.823Z",
  "tag": "nut.",
  "name": "nut.city",
  "finalRank": 10,
  "players": [
    {
      "name": "wolsne",
      "steamId": "76561197960315263",
      "isLeader": false,
      "joinedAt": "2020-01-07T11:52:14.170Z"
    },
    {
      "name": "dave2",
      "steamId": "76561198012709756",
      "isLeader": false,
      "joinedAt": "2020-01-06T00:59:28.900Z"
    },
    {
      "name": "fyg",
      "steamId": "76561198027610614",
      "isLeader": false,
      "joinedAt": "2020-01-07T19:19:21.530Z"
    },
    {
      "name": "trux",
      "steamId": "76561198043922390",
      "isLeader": false,
      "joinedAt": "2020-01-10T00:00:38.640Z"
    },
    {
      "name": "dale",
      "steamId": "76561198044052830",
      "isLeader": false,
      "joinedAt": "2020-01-07T20:54:24.816Z"
    },
    {
      "name": "pfart",
      "steamId": "76561198064534447",
      "isLeader": false,
      "joinedAt": "2020-01-07T18:27:51.870Z"
    },
    {
      "name": "tide*",
      "steamId": "76561198092224185",
      "isLeader": false,
      "joinedAt": "2020-01-21T00:39:53.476Z"
    },
    {
      "name": "Captain Zidgel",
      "steamId": "76561198098770013",
      "isLeader": true,
      "joinedAt": "2020-01-07T11:52:14.640Z"
    },
    {
      "name": "tsar",
      "steamId": "76561198116072296",
      "isLeader": true,
      "joinedAt": "2020-01-06T00:37:16.266Z"
    },
    {
      "name": "Connie",
      "steamId": "76561198124589076",
      "isLeader": false,
      "joinedAt": "2020-01-14T23:01:57.030Z"
    },
    {
      "name": "pug vibin to geico 15 minutes c",
      "steamId": "76561198128213108",
      "isLeader": false,
      "joinedAt": "2020-01-17T04:09:51.673Z"
    }
  ]
}`
	var expected Team
	err := json.Unmarshal([]byte(str), &expected)
	require.NoError(t, err, "Shouldn't get error unmarshalling expected json")

	got, err := r.GetTeam(5979)
	require.NoError(t, err, "Shouldn't get error querying for team")
	require.Equal(t, expected, got, "Should have equal team structs")

	got, err = r.GetTeam(111111)
	require.NoError(t, err, "Should get no error querying for 404")
	require.Equal(t, Team{}, got, "Should get empty object for 404")
}

func TestGetSeason(t *testing.T) {
	str := `{
  "name": "P7 Season 9",
  "formatName": null,
  "regionName": null,
  "maps": [
    "koth_product_rcx",
    "pl_vigil_rc8",
    "koth_synthetic_rc6a",
    "koth_cascade_v2_b5",
    "pl_upward",
    "koth_cascade_v2_b6",
    "pl_vigil_rc6",
    "koth_synthetic_rc2"
  ],
  "participatingTeams": [
    8317,
    8249,
    8250,
    8251,
    8252,
    8253,
    8254,
    8255,
    8258,
    8259,
    8260,
    8261,
    8263,
    8265,
    8266,
    8267,
    8268,
    8269,
    8270,
    8271,
    8272,
    8273,
    8274,
    8275,
    8276,
    8277,
    8278,
    8279,
    8280,
    8281,
    8282,
    8283,
    8284,
    8285,
    8286,
    8287,
    8289,
    8290,
    8291,
    8292,
    8293,
    8294,
    8295,
    8296,
    8297,
    8298,
    8299,
    8300,
    8301,
    8302,
    8303,
    8304,
    8305,
    8306,
    8307,
    8308,
    8309,
    8310,
    8311,
    8312,
    8313,
    8314,
    8315,
    8316,
    8318,
    8319,
    8320,
    8321,
    8322,
    8323,
    8324,
    8325,
    8327,
    8328,
    8329,
    8330,
    8331,
    8332,
    8333,
    8334,
    8335,
    8336,
    8337,
    8338,
    8339,
    8340,
    8341,
    8342,
    8343,
    8344,
    8345,
    8346,
    8348,
    8349,
    8350,
    8351,
    8352,
    8353,
    8354,
    8355,
    8356,
    8357,
    8358,
    8359,
    8360,
    8361,
    8362,
    8366,
    8367,
    8368,
    8369,
    8370,
    8371,
    8372,
    8373,
    8374,
    8375,
    8376,
    8377,
    8378,
    8379,
    8380,
    8640,
    8347,
    8264,
    8288,
    8326,
    8262
  ],
  "matchesPlayedDuringSeason": [
    13004,
    13005,
    13008,
    13009,
    13016,
    13017,
    13018,
    13019,
    13022,
    13025,
    13032,
    13033,
    13034,
    13035,
    13036,
    13037,
    13038,
    13039,
    13040,
    13041,
    13042,
    13043,
    13044,
    13045,
    13046,
    13047,
    13048,
    13049,
    13050,
    13051,
    13052,
    13053,
    13054,
    13055,
    13056,
    13057,
    13058,
    13059,
    13060,
    13061,
    13062,
    13063,
    13064,
    13065,
    13066,
    13067,
    13068,
    13069,
    13070,
    13071,
    13072,
    13073,
    13074,
    13075,
    13076,
    13077,
    13078,
    13079,
    13080,
    13082,
    13083,
    13084,
    13086,
    13087,
    13089,
    13091,
    13092,
    13093,
    13094,
    13095,
    13097,
    13099,
    13100,
    13101,
    13103,
    13105,
    13107,
    13108,
    13109,
    13110,
    13112,
    13114,
    13115,
    13116,
    13118,
    13119,
    13120,
    13122,
    13123,
    13124,
    13125,
    13126,
    13127,
    13128,
    13129,
    13130,
    13131,
    13132,
    13133,
    13134,
    13135,
    13136,
    13137,
    13138,
    13139,
    13140,
    13141,
    13142,
    13143,
    13145,
    13146,
    13147,
    13150,
    13152,
    13153,
    13154,
    13155,
    13156,
    13157,
    13158,
    13159,
    13160,
    13161,
    13162,
    13163,
    13164,
    13165,
    13166,
    13167,
    13168,
    13169,
    13170,
    13171,
    13172,
    13173,
    13174,
    13175,
    13176,
    13177,
    13178,
    13179,
    13182,
    13183,
    13184,
    13185,
    13186,
    13187,
    13188,
    13189,
    13190,
    13191,
    13192,
    13193,
    13194,
    13195,
    13196,
    13198,
    13199,
    13201,
    13202,
    13203,
    13205,
    13207,
    13208,
    13210,
    13211,
    13212,
    13213,
    13214,
    13215,
    13216,
    13217,
    13218,
    13219,
    13220,
    13221,
    13222,
    13223,
    13224,
    13225,
    13230,
    13232,
    13233,
    13234,
    13235,
    13236,
    13237,
    13238,
    13239,
    13240,
    13241,
    13242,
    13243,
    13244,
    13245,
    13246,
    13247,
    13248,
    13249,
    13250,
    13251,
    13252,
    13253,
    13255,
    13256,
    13257,
    13258,
    13259,
    13260,
    13276,
    13277,
    13278,
    13279,
    13280,
    13281,
    13282,
    13283,
    13284,
    13285,
    13732,
    13733,
    13734,
    13735,
    13736,
    13737,
    13738,
    13739,
    13740,
    13741,
    13742,
    13743,
    13744,
    13745,
    13746,
    13747,
    13748,
    13749,
    13750,
    13751,
    13752,
    13753,
    13754,
    13755,
    13756,
    13757,
    13795,
    13796,
    13995,
    13996,
    13997,
    13998,
    13999,
    14000,
    14001,
    14002,
    14003,
    14004,
    14005,
    14006,
    14007,
    14008,
    14009,
    14010,
    14011,
    14012,
    14013,
    14014,
    14015,
    14016,
    14017,
    14018,
    14019,
    14020,
    14021,
    14022
  ]
}`
	var expected Season
	err := json.Unmarshal([]byte(str), &expected)
	require.NoError(t, err, "Shouldn't get error unmarshalling expected json")

	got, err := r.GetSeason(107)
	require.NoError(t, err, "Shouldn't get error querying for season")
	require.Equal(t, expected, got, "Should have equal season structs")

	got, err = r.GetSeason(11111)
	require.NoError(t, err)
	require.Equal(t, Season{}, got, "Should get empty object for 404")
}

func TestSearchPlayers(t *testing.T) {
	_, err := r.SearchPlayers("a", 10, 10)
	require.EqualError(t, err, "Length of alias must be at least 2")

	str := `{
  "results": [
    "76561198098770013"
  ],
  "count": 1,
  "totalHitCount": 1
}`
	var expected SearchResults
	json.Unmarshal([]byte(str), &expected)
	got, err := r.SearchPlayers("Zidgel", 100, 0)
	require.NoError(t, err)
	require.Equal(t, expected, got)

	got, err = r.SearchPlayers("No one has this alias!", 1, 0)
	require.NoError(t, err)
	require.Equal(t, len(SearchResults{}.Results), 0, "Should get empty results for 404")
	require.Equal(t, SearchResults{Results: make([]string, 0)}, got, "Should get empty object for no matches")
}

func TestBulkPlayers(t *testing.T) {
	str := `[
  {
    "steamId": "76561198098770013",
    "avatar": "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/81/8148c37b434814fb7a4bc175a608c1353b4d0a11_full.jpg",
    "name": "Captain Zidgel",
    "updatedAt": "2023-02-12T21:48:27.196Z",
    "status": {
      "isVerified": false,
      "isBanned": false,
      "isOnProbation": false
    },
    "banInformation": null,
    "currentTeams": {
      "sixes": null,
      "highlander": null,
      "prolander": null
    }
  },
  {
    "steamId": "76561197970669109",
    "avatar": "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/8d/8dbbf447da3ccdb8982b9c6b9257d75fb940f9a1_full.jpg",
    "name": "b4nny",
    "updatedAt": "2022-12-06T07:44:15.356Z",
    "status": {
      "isVerified": true,
      "isBanned": false,
      "isOnProbation": false
    },
    "banInformation": null,
    "currentTeams": {
      "sixes": {
        "id": 11088,
        "tag": "FROYO",
        "name": "froyotech",
        "status": "Ready",
        "seasonId": 133,
        "divisionId": 809,
        "divisionName": "Invite"
      },
      "highlander": null,
      "prolander": null
    }
  }
]`
	p, err := r.BulkPlayers([]string{"76561198292350104"})
	require.NoError(t, err)
	require.Len(t, p, 0, "Should get no results for non-rgl player")

	p, err = r.BulkPlayers([]string{"765611980987700133"})
	require.NoError(t, err)
	require.Len(t, p, 0, "Should get no results for invalid ID")

	p, err = r.BulkPlayers([]string{"76561198098770013", "76561197970669109", "765611980987700133"}) //Last ID is invalid
	require.NoError(t, err)
	var expected []Player
	json.Unmarshal([]byte(str), &expected)
	require.Equal(t, expected, p, "Should get accurate results for a bulk search")
}

func TestGetMatch(t *testing.T) {
	str := `{
  "matchId": 5256,
  "seasonName": "Sixes S2",
  "divName": "Intermediate",
  "seasonId": 67,
  "matchDate": "2020-01-15T03:30:00.000Z",
  "matchName": "Week 1A",
  "winner": 5979,
  "teams": [
    {
      "teamName": "nut.city",
      "teamTag": "nut.",
      "teamId": 5979,
      "isHome": false,
      "points": "2.75"
    },
    {
      "teamName": "Sunny",
      "teamTag": "s.",
      "teamId": 5819,
      "isHome": false,
      "points": "0.25"
    }
  ],
  "maps": [
    {
      "mapName": "cp_snakewater_final1",
      "homeScore": 1,
      "awayScore": 5
    }
  ]
}`

	var expected Match
	json.Unmarshal([]byte(str), &expected)

	m, err := r.GetMatch(5256)
	require.NoError(t, err)
	require.Equal(t, expected, m)

	m, err = r.GetMatch(555555)
	require.NoError(t, err)
	require.Equal(t, Match{}, m, "Should get empty object for 404")
}

func TestSearchTeam(t *testing.T) {
	str := `{
  "results": [
    "42",
    "83",
    "1142",
    "2460",
    "2512",
    "3558",
    "3754",
    "3755",
    "5016",
    "5754",
    "5829",
    "6116",
    "6332",
    "6944",
    "7416",
    "7835",
    "7836",
    "8551",
    "8585",
    "9031",
    "9142",
    "9187",
    "9846",
    "10269",
    "10481"
  ],
  "count": 25,
  "totalHitCount": 27
}`
	var expected SearchResults
	json.Unmarshal([]byte(str), &expected)

	results, err := r.SearchTeams("froyo", 25, 0)
	require.NoError(t, err)
	require.Equal(t, expected, results, "Should get results for valid request")

	results, err = r.SearchTeams("No one has this team name!", 1, 1)
	require.NoError(t, err)
	require.Equal(t, SearchResults{Results: make([]string, 0)}, results, "Should get empty results for 404")
}

func TestGetPlayerTeamHistory(t *testing.T) {
	str := `[
  {
    "formatId": 3,
    "formatName": "Sixes",
    "regionId": 40,
    "regionName": "NA Sixes",
    "seasonId": 67,
    "seasonName": "Sixes S2",
    "startedAt": "2020-01-07T11:52:14.640Z",
    "divisionId": 363,
    "divisionName": "Intermediate",
    "leftAt": "2020-04-03T00:00:00.000Z",
    "teamName": "nut.city",
    "teamTag": "nut.",
    "teamId": 5979,
    "stats": {
      "wins": 9,
      "winsWithout": 2,
      "loses": 7,
      "losesWithout": 4,
      "gamesPlayed": 16,
      "gamesWithout": 6
    }
  }
]`
	var expected []PlayerTeamHistory
	json.Unmarshal([]byte(str), &expected)

	results, err := r.GetPlayerTeamHistory("76561198098770013")
	require.NoError(t, err)
	require.Equal(t, expected, results, "Should get results for valid request")

	results, err = r.GetPlayerTeamHistory("765611980987700133")
	require.NoError(t, err)
	require.Equal(t, make([]PlayerTeamHistory, 0), results, "Should get empty slice for invalid request")
}

func TestGetBans(t *testing.T) {
	// impossible to compare to a static string here since new bans are always incoming

	results, err := r.GetBans(100, 0)
	require.NoError(t, err)
	require.True(t, len(results) == 100, "Should get 100 values")
	require.True(t, results[0] != BulkBan{}, "Values shouldn't be empty")

	results, err = r.GetBans(1, 999999)
	require.NoError(t, err)
	require.True(t, len(results) == 0)
}
