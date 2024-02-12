package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Define structs to unmarshal JSON responses
type LastFmTrack struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Image  []struct {
		URL string `json:"#text"`
	} `json:"image"`
}

type MusixmatchLyrics struct {
	Lyrics string `json:"lyrics_body"`
}

type ArtistInfo struct {
	Name      string   `json:"name"`
	ImageURL  string   `json:"image_url"`
	SimilarTo []string `json:"similar_to"`
}

func main() {
	http.HandleFunc("/toptrack", topTrackHandler)
	http.ListenAndServe(":8099", nil)
}

func topTrackHandler(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")

	// Fetch top track from Last.fm
	topTrack, err := fetchTopTrack(region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch lyrics from Musixmatch
	lyrics, err := fetchLyrics(topTrack.Name, topTrack.Artist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch artist info and image
	artistInfo, err := fetchArtistInfo(topTrack.Artist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Combine data and send response
	response := struct {
		TopTrack LastFmTrack `json:"top_track"`
		Lyrics   string      `json:"lyrics"`
		Artist   ArtistInfo  `json:"artist_info"`
	}{
		TopTrack: topTrack,
		Lyrics:   lyrics,
		Artist:   artistInfo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func fetchTopTrack(region string) (LastFmTrack, error) {
	apiKey := "YOUR_SERPAPI_API_KEY"
	url := fmt.Sprintf("http://ws.audioscrobbler.com/2.0/?method=geo.gettoptracks&country=%s&api_key=%s&format=json", region, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return LastFmTrack{}, err
	}
	defer resp.Body.Close()

	var data struct {
		Tracks struct {
			Track []LastFmTrack `json:"track"`
		} `json:"tracks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return LastFmTrack{}, err
	}

	if len(data.Tracks.Track) == 0 {
		return LastFmTrack{}, fmt.Errorf("no top tracks found for the region %s", region)
	}

	return data.Tracks.Track[0], nil // Return the top track
}

func fetchLyrics(trackName, artistName string) (string, error) {
	apiKey := "YOUR_SERPAPI_API_KEY"
	url := fmt.Sprintf("https://api.musixmatch.com/ws/1.1/matcher.lyrics.get?format=json&apikey=%s&q_track=%s&q_artist=%s", apiKey, url.QueryEscape(trackName), url.QueryEscape(artistName))

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Message struct {
			Body struct {
				Lyrics MusixmatchLyrics `json:"lyrics"`
			} `json:"body"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if data.Message.Body.Lyrics.Lyrics == "" {
		return "", fmt.Errorf("lyrics not found for the track %s by %s", trackName, artistName)
	}

	return data.Message.Body.Lyrics.Lyrics, nil
}

func fetchArtistInfo(artistName string) (ArtistInfo, error) {
	apiKey := "YOUR_SERPAPI_API_KEY"
	search := google_search.NewGoogleSearch(apiKey)

	params := map[string]string{
		"q":       artistName,
		"tbm":     "isch", // Image search
		"num":     "1",    // Number of results
		"safe":    "active",
		"api_key": apiKey,
	}

	resp, err := search.JSON(params)
	if err != nil {
		return ArtistInfo{}, err
	}

	// Extract artist info and image URL from the response
	artistInfo := ArtistInfo{
		Name:     artistName,
		ImageURL: resp["images_results"].([]interface{})[0].(map[string]interface{})["original"].(string),
	}

	return artistInfo, nil
}
