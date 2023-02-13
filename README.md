# RGL

##### Note that RGL's API is in alpha right now and subject to change. This library could break at any moment.

Reference:  
[![Go Reference](https://pkg.go.dev/badge/github.com/captainzidgel/rgl.svg)](https://pkg.go.dev/github.com/captainzidgel/rgl)
[![Go Report Card](https://goreportcard.com/badge/github.com/captainzidgel/rgl)](https://goreportcard.com/report/github.com/captainzidgel/rgl)

Tips:  
Create an RGL object: `r := rgl.DefaultRateLimit()`  
Then do stuff. `player, err := r.GetPlayer("steam64")`  
404/Not Found errors do not return errors, they return zero values where 404 represents an expected "no results for your query".  
I made the decision to avoid nil values where possible, so check the zero values carefully.

A type without slices in it can be compared to like `player != Player{}`, but when types include slices, the slice has to be made: `results != SearchResults{Results: make([]string, 0)}`. This is verbose so you can use something like `len(results.Results > 0)` , or just test a single field. `results.Count > 0` is readable for SearchResults, but for something like Team you'll probably want to check `team.Id > 0`. Just take a look at the types in the reference.

You can use your own implementation of the rate.Limiter object by setting `r.rl`, or you can set it to nil and ratelimit your requests yourself.

Some fields are time strings. Convert to time.Time with `t := rgl.ToGoTime(ban.Ends)`