package discord

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
)

const (
	sendTimeout = 10 * time.Second
)

var (
	// Pre-compiled regexes for resolveDiscordRefs (avoid re-compiling per call)
	channelRefRe = regexp.MustCompile(`<#(\d+)>`)
	msgLinkRe    = regexp.MustCompile(`https://(?:discord\.com|discordapp\.com)/channels/(\d+)/(\d+)/(\d+)`)
)

// appendContent safely appends content to existing text
func appendContent(content, suffix string) string {
	if content == "" {
		return suffix
	}
	return content + "\n" + suffix
}

func applyDiscordProxy(session *discordgo.Session, proxyAddr string) error {
	var proxyFunc func(*http.Request) (*url.URL, error)
	if proxyAddr != "" {
		proxyURL, err := url.Parse(proxyAddr)
		if err != nil {
			return fmt.Errorf("invalid discord proxy URL %q: %w", proxyAddr, err)
		}
		proxyFunc = http.ProxyURL(proxyURL)
	} else if os.Getenv("HTTP_PROXY") != "" || os.Getenv("HTTPS_PROXY") != "" {
		proxyFunc = http.ProxyFromEnvironment
	}

	if proxyFunc == nil {
		return nil
	}

	transport := &http.Transport{Proxy: proxyFunc}
	session.Client = &http.Client{
		Timeout:   sendTimeout,
		Transport: transport,
	}

	if session.Dialer != nil {
		dialerCopy := *session.Dialer
		dialerCopy.Proxy = proxyFunc
		session.Dialer = &dialerCopy
	} else {
		session.Dialer = &websocket.Dialer{Proxy: proxyFunc}
	}

	return nil
}
