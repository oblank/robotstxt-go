// The robots.txt Exclusion Protocol is implemented as specified in
// http://www.robotstxt.org/wc/robots.html
// with various extensions.
package robotstxt

import (
    "os"
    "strings"
)


type RobotsData struct {
    DefaultAgent string
    // private
    rules       []Rule
    allowAll    bool
    disallowAll bool
}

type Rule struct {
    Agent string
    Uri   string
    Allow bool
}

var AllowAll = &RobotsData{allowAll: true}
var DisallowAll = &RobotsData{disallowAll: true}


func FromResponse(statusCode int, body string, print_errors bool) (*RobotsData, os.Error) {
    switch {
    case statusCode == 404:
        return AllowAll, nil
    case statusCode == 401 || statusCode == 403:
        return DisallowAll, nil
    case statusCode >= 200 && statusCode < 300:
        return FromString(body, print_errors)
    }
    // Conservative disallow all default
    return DisallowAll, nil
}

func FromString(body string, print_errors bool) (r *RobotsData, err os.Error) {
    // special case (probably not worth optimization?)
    trimmed := strings.TrimSpace(body)
    if trimmed == "" {
        return AllowAll, nil
    }

    sc := NewByteScanner("string", false)
    sc.Quiet = !print_errors
    sc.Feed([]byte(body), true)
    var tokens []string
    tokens, err = sc.ScanAll()
    if err != nil {
        return nil, err
    }

    // special case worth optimization
    if len(tokens) == 0 {
        return AllowAll, nil
    }

    r = &RobotsData{}
    parser := NewParser(tokens)
    r.rules, err = parser.ParseAll()

    return r, err
}

func (r *RobotsData) Test(url string) (bool, os.Error) {
    if r.DefaultAgent == "" {
        return false, os.NewError("DefaultAgent is empty. You MUST set RobotsData.DefaultAgent to use Test method.")
    }
    return r.TestAgent(url, r.DefaultAgent)
}

func (r *RobotsData) TestAgent(url, agent string) (allow bool, err os.Error) {
    if r.allowAll {
        return true, nil
    }
    if r.disallowAll {
        return false, nil
    }

    // optimistic
    allow = true
    for _, rule := range r.rules {
        if rule.MatchAgent(agent) && rule.MatchUrl(url) {
            allow = rule.Allow
            // stop on first disallow as safety default
            // in absense of better algorithm
            if !rule.Allow {
                break
            }
        }
    }

    return allow, nil
}

func (rule *Rule) MatchAgent(agent string) bool {
    l_agent := strings.ToLower(agent)
    l_rule_agent := strings.ToLower(rule.Agent)
    return rule.Agent == "*" || strings.HasPrefix(l_agent, l_rule_agent)
}

func (rule *Rule) MatchUrl(url string) bool {
    return strings.HasPrefix(url, rule.Uri)
}

func (rule *Rule) String() string {
    allow_str := "Disallow"
    if rule.Allow {
        allow_str = "Allow"
    }
    return "<" + allow_str + " " + rule.Agent + " " + rule.Uri + ">"
}
