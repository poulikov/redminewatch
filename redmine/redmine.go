package redmine

import (
    "time"
    "net/http"
    "sort"
    "log"
    "encoding/json"
    "os"
    "os/user"
    "net/url"
)

var (
    ApiConfig Config
    ApiUrl string
)

func init() {
    if usr, err := user.Current(); err == nil {
        fl, err := os.Open(usr.HomeDir + "/.redminewatch.json")
        if err != nil {
            log.Fatal("Can't open ~/.redminewatch.json")
        }
        defer fl.Close()
        dec := json.NewDecoder(fl)
        dec.Decode(&ApiConfig)
    } else {
        log.Fatal("Can't get user home directory.")
    }

    u, err := url.Parse(ApiConfig.Url)
    if err != nil {
        log.Fatal(err)
    }
    q := u.Query()
    q.Set("key", ApiConfig.Key)
    for k,v := range ApiConfig.Filters {
        q.Set(k, v)
    }
    u.RawQuery = q.Encode()

    ApiUrl = u.String()
}

type Config struct {
    Url string
    Key string
    Filters map[string]string
}

type IssueFeed struct {
    Total_count int
    Issues []Issue
}

func (feed *IssueFeed) Len() int {
    return len(feed.Issues)
}

func (feed *IssueFeed) Swap(i, j int) {
    feed.Issues[i], feed.Issues[j] = feed.Issues[j], feed.Issues[i]
}

func (feed *IssueFeed) Less(i, j int) bool {
    return feed.Issues[i].Updated_on > feed.Issues[j].Updated_on
}

func (feed *IssueFeed) OlderThan(tm *time.Time) []Issue {
    sort.Sort(feed)

    var index int32 = 0

    for idx := range feed.Issues {
        if feed.Issues[idx].OlderThan(tm) {
            break
        }
        index++
    }

    return feed.Issues[:index]
}

func LoadTasks() (*IssueFeed, error) {
    resp, err := http.Get(ApiUrl)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    dec := json.NewDecoder(resp.Body)

    tasks := new(IssueFeed)
    dec.Decode(tasks)

    sort.Sort(tasks)

    return tasks, err
}

type Issue struct {
    Id int
    Custom_fields []IssueCustomField
    Assigned_to IssueField
    Project IssueField
    Description string
    Start_date string
    Done_ratio int
    Author IssueField
    Tracker IssueField
    Created_on string
    Priority IssueField
    Subject string
    Status IssueField
    Updated_on string
}

func (iss *Issue) OlderThan(tm *time.Time) bool {
    tm2, err := time.Parse(time.RFC3339, iss.Updated_on)
    if err != nil {
        log.Println(err)
        return false
    }

    if tm2.Before(*tm) {
        return true
    }
    return false
}

func (iss *Issue) LastUpdate() *time.Time {
    tm, err := time.Parse(time.RFC3339, iss.Updated_on);
    if err != nil {
        log.Println(err)
        return nil
    }
    localtm := tm.Local()

    return &localtm
}

type IssueCustomField struct {
    Id int
    Value string
    Name string
}

type IssueField struct {
    Id int
    Name string
}

