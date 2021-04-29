package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	apiURL   string
	clientID string
	client   *http.Client
}

const defaultTwitchURL = "https://api.twitch.tv/kraken"
const twitchV5AcceptHeader = "application/vnd.twitchtv.v5+json"

type ClientOption func(*Client)

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		apiURL:   defaultTwitchURL,
		clientID: "",
	}

	for _, o := range opts {
		o(c)
	}

	if c.client == nil {
		// Better Defaults
		c.client = &http.Client{}
	}

	return c
}

func WithClientID(clientID string) ClientOption {
	return func(c *Client) {
		c.clientID = clientID
	}
}

func WithAPIURL(apiurl string) ClientOption {
	return func(c *Client) {
		c.apiURL = apiurl
	}
}

func (c *Client) annotateRequest(req *http.Request) (*http.Request, error) {
	req.Header.Add("Accept", twitchV5AcceptHeader)
	req.Header.Add("Client-ID", c.clientID)
	return req, nil
}

type Stream struct {
	ID          int       `json:"_id"`
	Game        string    `json:"game"`
	Preview     Thumbnail `json:"preview"`
	VideoHeight int       `json:"video_height"`
	Viewers     int       `json:"viewers"`
}

type Thumbnail struct {
	Template string `json:"template"`
}

type GetStreamsResponse struct {
	Streams []Stream `json:"streams"`
}

func (c *Client) GetTopStreams(ctx context.Context, game string, limit int) ([]Stream, error) {
	streams := make([]Stream, limit)
	for i := 0; i < limit/100; i++ {
		offset := i * 100
		rlimit := limit - offset
		if rlimit > 100 {
			rlimit = 100
		}

		url := fmt.Sprintf("%s/streams?limit=%d&offset=%d", c.apiURL, rlimit, offset)
		if game != "" {
			url += "&game=" + game
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req, err = c.annotateRequest(req)
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var content GetStreamsResponse
		err = json.NewDecoder(resp.Body).Decode(&content)
		if err != nil {
			return nil, err
		}
		for i := range content.Streams {
			streams[i+offset] = content.Streams[i]
		}
	}

	return streams, nil
}

func GetURLForStreamShot(stream *Stream) (string, error) {
	videoHeight := stream.VideoHeight
	videoWidth := videoHeight * 16 / 9

	url := stream.Preview.Template
	url = strings.Replace(url, "{width}", strconv.Itoa(videoWidth), 1)
	url = strings.Replace(url, "{height}", strconv.Itoa(videoHeight), 1)
	return url, nil
}
