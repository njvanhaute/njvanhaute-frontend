package main

import (
	"fmt"
	"net/http"
	"time"

	"frontend.njvanhaute.com/internal/models"
)

type Tune struct {
	ID            int64     `json:"id"`             // Unique integer ID for the tune
	CreatedAt     time.Time `json:"-"`              // Timestamp for when the tune is added to our database
	Title         string    `json:"title"`          // Tune title
	Styles        []string  `json:"styles"`         // Slice of styles for the tune (Bluegrass, old time, Irish, etc.)
	Keys          []string  `json:"keys"`           // Slice of keys for the tune (ex: A major, G minor)
	TimeSignature string    `json:"time_signature"` // Tune time signature
	Structure     string    `json:"structure"`      // Tune structure (ex: AABA)
	HasLyrics     bool      `json:"has_lyrics"`     // Whether or not the tune has lyrics
}

type TuneEnvelope struct {
	Tune Tune `json:"tune"`
}

func (app *application) Insert(tune Tune) (int, error) {
	return 0, nil
}

func (app *application) GetTune(id int, r *http.Request) (Tune, error) {
	endpoint := fmt.Sprintf("/v1/tunes/%d", id)

	req, err := http.NewRequest(http.MethodGet, app.buildURL(endpoint), nil)
	if err != nil {
		return Tune{}, err
	}

	token := app.sessionManager.Get(r.Context(), "authenticatedUserToken")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.httpClient.Do(req)

	app.logger.Info("Req", "req", req)
	app.logger.Info("Resp", "resp", resp)

	if err != nil {
		return Tune{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return Tune{}, models.ErrNoRecord
	}

	var tuneEnvelope TuneEnvelope

	err = app.readJSON(resp, &tuneEnvelope)
	if err != nil {
		return Tune{}, err
	}

	return tuneEnvelope.Tune, nil
}

func (app *application) Latest() ([]Tune, error) {
	return nil, nil
}
